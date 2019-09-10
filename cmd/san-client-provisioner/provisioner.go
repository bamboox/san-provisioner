package main

import (
	"flag"
	"fmt"
	log "github.com/golang/glog"
	"github.com/kubernetes-sigs/san-client/fc"
	"k8s.io/api/core/v1"
	storage "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	volumehelpers "k8s.io/cloud-provider/volume/helpers"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
	"os"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
	"strconv"
	"strings"
)

const (
	ThinRateAnnotation    = "volume.beta.kubernetes.io/mount-options"
	MountOptionAnnotation = "volume.beta.kubernetes.io/mkfs-fstype"
	MkfsFsTypeAnnotation  = "volume.beta.kubernetes.io/thin-rate"
	DiskNameAnnotation    = "volume.beta.kubernetes.io/disk-name"

	EventComponent = "san-client-provisioner"
)

type sanProvisioner struct {
	client        kubernetes.Interface
	eventRecorder record.EventRecorder
}

// NewSanProvisioner creates a new san provisioner
func NewSanProvisioner(client kubernetes.Interface) controller.Provisioner {

	//TODO: remove this
	v1.AddToScheme(scheme.Scheme)
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: EventComponent})

	return &sanProvisioner{
		client:        client,
		eventRecorder: eventRecorder,
	}
}

var _ controller.Provisioner = &sanProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *sanProvisioner) Provision(options controller.ProvisionOptions) (*v1.PersistentVolume, error) {
	log.V(5).Infof("Provision options: %+v", options)
	// step1. create disk
	volSizeBytes := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	capacity, err := volumehelpers.RoundUpToGiBInt(volSizeBytes)
	if err != nil {
		return nil, err
	}
	params := options.StorageClass.Parameters
	fc := fc.NewFcClient(params["user"], params["password"], params["sanserver"], 22, params["sanCliPipelineEndpoint"])
	thinRate := ""
	if _, ok := options.PVC.Annotations[ThinRateAnnotation]; ok {
		thinRate = options.PVC.Annotations[ThinRateAnnotation]
	}
	volumeName := p.genVolumeName(options.PVName, options.PVC)
	createErr := fc.CreateVolume(volumeName, strconv.Itoa(capacity), thinRate)
	if createErr != nil {
		return nil, createErr
	}
	// step2. GetWWIdByVolume
	wwid, pvcErr := fc.GetWWIdByVolume(volumeName)
	if pvcErr != nil {
		return nil, pvcErr
	}
	log.V(5).Infof("GetWWIdByVolume wwid: %s", wwid)
	// step3. mapHostForVolume
	// TODO bind node
	if options.SelectedNode != nil {
		nodeName := options.SelectedNode.Name
		if _, ok := options.SelectedNode.Labels["kubernetes.io/hostname"]; ok {
			nodeName = options.SelectedNode.Labels["kubernetes.io/hostname"]
		}
		log.Infof("SelectedNode nodeName: %s", nodeName)
	}

	err = fc.MapAllHostForVolume(volumeName)
	if err != nil {
		return nil, err
	}

	// step4. removeAbnormalDevice

	//get all k8s nodeAddress
	nodes, err := p.client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("get nodes err :%v", err)
		return nil, err
	}
	nodesIp, err := getAllNodeIP(nodes)
	if err != nil || nodesIp == nil {
		log.Errorf("getAllNodeIP nodesIp: %+v, err :%v", nodesIp, err)
		return nil, err
	}

	err = fc.RemoveAbnormalDevice(nodesIp, volumeName, wwid)
	if err != nil {
		log.Errorf("RemoveAbnormalDevice err :%v", err)
		return nil, err
	}

	// step5. create pv
	// fsType default ext4 and  mountOptions
	fsType := "ext4"
	var mountOptions []string
	if options.PVC.Annotations != nil {
		if ft, ok := options.PVC.Annotations[MkfsFsTypeAnnotation]; ok {
			fsType = ft
		}
		if mo, ok := options.PVC.Annotations[MountOptionAnnotation]; ok {
			mountOptions = strings.Split(mo, ",")
		}
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				DiskNameAnnotation: volumeName,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FC: &v1.FCVolumeSource{
					WWIDs:  []string{wwid},
					FSType: fsType,
				},
			},
			MountOptions: mountOptions,
		},
	}
	// check disk is normal mapper
	log.V(5).Infof("provision pv success")
	return pv, nil
}

func (p *sanProvisioner) genVolumeName(pvName string, pvc *v1.PersistentVolumeClaim) string {
	volumeName := fmt.Sprintf("%s_%s", pvc.Name, pvc.Namespace)
	// FC SAN disk name max length 63
	if len(volumeName) > 63 {
		p.eventRecorder.Event(pvc, v1.EventTypeWarning, "genVolumeName", fmt.Sprintf("volumeName max length 63, this use pv name, vn: %s", pvName))
		return pvName
	}
	return volumeName
}

func getAllNodeIP(nodes *v1.NodeList) (nodeIP map[string]string, err error) {
	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("nodes is nill")
	}
	nodeIP = make(map[string]string)
	for _, node := range nodes.Items {
		var name, ip string

		name = node.Name
		for _, addr := range node.Status.Addresses {
			log.V(5).Infof("node: %s, addr: %+v", name, addr)
			if addr.Type == v1.NodeInternalIP {
				ip = addr.Address
				break
			}
		}
		if len(ip) == 0 {
			return nil, fmt.Errorf("node: %s, not found NodeExternalIP", name)
		}
		// HACK: san host name
		nodeIP[fmt.Sprintf("h_%s", name)] = ip
	}
	return
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *sanProvisioner) Delete(volume *v1.PersistentVolume) error {
	log.V(5).Infof("Delete volume: %+v", volume)

	class, err := p.getClassForVolume(volume)
	if err != nil {
		return err
	}
	params := class.Parameters

	volumeName := ""
	if dn, ok := volume.Annotations[DiskNameAnnotation]; ok {
		volumeName = dn
	} else {
		return fmt.Errorf("pv: %s, annotation not found disk name ", volume.Name)
	}

	fc := fc.NewFcClient(params["user"], params["password"], params["sanserver"], 22, params["sanCliPipelineEndpoint"])
	log.V(5).Infof("DeleteVolume: begin delete san: %s, %s, %s, %s", volumeName, params["user"], params["password"], params["sanserver"])
	// step1. RemoveAllMapping
	err = fc.RemoveAllMapping(volumeName)
	if err != nil {
		return err
	}
	// step2. DeleteVolume
	err = fc.DeleteVolume(volumeName)
	if err != nil {
		return err
	}
	log.V(5).Infof("provision delete pv success")
	return nil
}

// getClassForVolume returns StorageClass
func (p *sanProvisioner) getClassForVolume(pv *v1.PersistentVolume) (*storage.StorageClass, error) {

	className := helper.GetPersistentVolumeClass(pv)
	if className == "" {
		return nil, fmt.Errorf("Volume has no storage class")
	}
	class, err := p.client.StorageV1().StorageClasses().Get(className, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return class, nil
}

func main() {
	flag.Parse()
	flag.Set("logtostderr", "true")
	// Create an InClusterConfig and use it to create a client for the controller
	// to use to communicate with Kubernetes
	log.V(5).Infof("san provisioner start")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	provisionerName := os.Getenv("PROVISIONER_NAME")
	if provisionerName == "" {
		log.Fatalf("environment variable %s is not set! Please set it.", provisionerName)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		log.Fatalf("Error getting server version: %v", err)
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	sanProvisioner := NewSanProvisioner(clientset)

	// Start the provision controller which will dynamically provision hostPath
	pc := controller.NewProvisionController(clientset, provisionerName, sanProvisioner, serverVersion.GitVersion)
	//
	pc.Run(wait.NeverStop)
}
