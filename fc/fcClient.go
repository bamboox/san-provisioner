package fc

import (
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"regexp"
	"strings"
	"time"
)

const (
	// cmd ignore error
	//volume is already mappe
	CmdMkVDiskHostMapIgnoreAndErr = "volume is already mapped"
)

type FcClient struct {
	sanCliPipelineEndpoint string
	user                   string
	password               string
	host                   string
	port                   int
}

func commonErrFmtNLog(ctxKey, out, errInfo string, err error) error {
	errOut := fmt.Errorf("%s failed; out: [%s], err: [%v], errInfo: [%s]", ctxKey, out, err, errInfo)
	log.Errorf("%v", errOut)
	return errOut
}

func NewFcClient(user, password, host string, port int, sanCliPipelineEndpoint string) (fc *FcClient) {
	return &FcClient{
		user:                   user,
		password:               password,
		host:                   host,
		port:                   port,
		sanCliPipelineEndpoint: sanCliPipelineEndpoint,
	}
}

func (fcClient *FcClient) runSsh(cmd string, tags ...string) (outInfo, errInfo string, err error) {
	tag := strings.Join(tags, "|")

	start := time.Now()
	log.Infof("runSsh-tag[%s]SendSanCMD start time::%v", tag, start)

	outStr, errInfo, err := RunSSHCMD(cmd, fcClient.host, fcClient.user, "runSsh")

	end := time.Now()
	log.Infof("runSsh-tag[%s]RunSSHLocalCMD end time:%v", tag, end)
	log.Infof("runSsh-tag[%s]RunSSHLocalCMD cost time:%v(s)", tag, end.Sub(start).Seconds())
	log.Infof("runSsh-tag[%s] Stdout: %s, Stderr: %s, ErrInfo: %v", tag, outStr, errInfo, err)

	return outStr, errInfo, nil
}

func (fc *FcClient) CreateVolume(volumeName string, size string, thinRate string) error {
	createVolumeCmd := ""
	if thinRate == "" {
		createVolumeCmd = getCreateVolumeCmd(volumeName, size)
	} else {
		createVolumeCmd = getCreateThinVolumeCmd(volumeName, size, thinRate)
	}
	out, errInfo, err := fc.runSsh(createVolumeCmd, volumeName)
	if (err != nil || errInfo != "") && !strings.Contains(errInfo, " the object already exists") {
		return commonErrFmtNLog(fmt.Sprintf("createVolume [%s]", volumeName), out, errInfo, err)
	}
	log.Infof("createVolume [%s] succeed out: [%s]", volumeName, out)
	return nil
}

func getCreateVolumeCmd(volumeName string, size string) string {
	return fmt.Sprintf("mkvdisk -name %s -mdiskgrp 0 -size %s -unit gb -nofmtdisk", volumeName, size)
}

func getCreateThinVolumeCmd(volumeName, size, thinRate string) string {
	return fmt.Sprintf("mkvdisk -name %s -autoexpand -cache readwrite -grainsize 256 -mdiskgrp 0  -rsize %s -size %s -unit gb -warning 80% -nofmtdisk", volumeName, thinRate, size)
}

func getMapHostForVolumeCmd(volumeName string, hostId string) string {
	return fmt.Sprintf("mkvdiskhostmap -force -host %s %s", hostId, volumeName)
}

func getHostIdCmd() string {
	return "lshost -delim '|'"
}

func (fc *FcClient) getHostIdByNodeName(nodeIp string) (hostId string, err error) {
	ret := ""
	out, errInfo, err := fc.runSsh(getHostIdCmd())
	if err != nil || errInfo != "" {
		return ret, commonErrFmtNLog(fmt.Sprintf("getHostIdByNodeName [%s]", nodeIp), out, errInfo, err)
	}
	log.Infof("getHostIdByNodeName succeed; nodeIp: [%s], out: [%s]", nodeIp, out)

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, nodeIp) {
			return strings.Split(line, "|")[0], nil
		}
	}
	return ret, commonErrFmtNLog(fmt.Sprintf("getHostIdByNodeName nodeIp: [%s]", nodeIp), out, errInfo, err)
}

func (fc *FcClient) getAllHostId() (hostIds []string, err error) {
	ret := make([]string, 0)
	out, errInfo, err := fc.runSsh(getHostIdCmd())
	if err != nil || errInfo != "" {
		return nil, commonErrFmtNLog("getAllHostId", out, errInfo, err)
	}
	log.Infof("getAllHostId succeed; out: [%s]", out)
	lines := strings.Split(out, "\n")
	for i := 1; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			continue
		}
		ret = append(ret, strings.Split(lines[i], "|")[0])
	}
	log.V(5).Infof("getAllHostId succeed; out: [%v]", ret)
	return ret, nil
}

func IsAlreadyMapped(out, errInfo string) bool {
	if strings.Contains(out, CmdMkVDiskHostMapIgnoreAndErr) || strings.Contains(errInfo, CmdMkVDiskHostMapIgnoreAndErr) {
		return true
	}
	return false
}
func (fc *FcClient) MapHostForVolume(volumeName, nodeName string, hostid string) error {
	if hostid == "" {
		hostIdTemp, err := fc.getHostIdByNodeName(nodeName)
		if err != nil {
			return err
		}
		hostid = hostIdTemp
	}

	mapHostCmd := getMapHostForVolumeCmd(volumeName, hostid)
	out, errInfo, err := fc.runSsh(mapHostCmd)
	if (err != nil || errInfo != "") && !IsAlreadyMapped(out, errInfo) {
		return commonErrFmtNLog(fmt.Sprintf("mapHostCmd: cmd[%s", mapHostCmd), out, errInfo, err)
	}
	log.Infof("mapHostForVolume %s succeed out: [%s]", volumeName, out)
	return nil
}

func (fc *FcClient) MapAllHostForVolume(volumeName string) error {
	hostIds, err := fc.getAllHostId()
	if err != nil {
		return err
	}
	for _, hostId := range hostIds {
		mapHostCmd := getMapHostForVolumeCmd(volumeName, hostId)
		out, errInfo, err := fc.runSsh(mapHostCmd)
		if (err != nil || errInfo != "") && !IsAlreadyMapped(out, errInfo) {
			return commonErrFmtNLog(fmt.Sprintf("mapHostCmd: cmd[%s", mapHostCmd), out, errInfo, err)
		}
		log.Infof("mapHostForVolume %s succeed out: [%s]", volumeName, out)
	}
	return nil
}

/*vdisk_UID
# example output
# id|name|IO_group_id|IO_group_name|status|mdisk_grp_id|mdisk_grp_name|capacity|type|LC_id|LC_name|RC_id|RC_name|vdisk_UID|lc_map_count|copy_count|fast_write_state|se_copy_count|RC_change|compressed_copy_count|parent_mdisk_grp_id|parent_mdisk_grp_name|formatting|encrypt|volume_id|volume_name|function|ica|ica_bypass|ica_pid
# 0|cx_test1|0|io_grp0|online|0|polardb-o|500.00GB|striped|||||60050767088080A26800000000000004|0|1|empty|0|no|0|0|polardb-
*/

func getVdiskUIDIndex(out string) int {
	lines := strings.Split(out, "\n")
	vdiskUIDIndex := 0
	for _, line := range lines {
		if strings.Contains(line, "vdisk_UID") {
			columes := strings.Split(line, "|")
			for index, colume := range columes {
				if colume == "vdisk_UID" {
					vdiskUIDIndex = index
				}
			}
		}
	}
	return vdiskUIDIndex
}

func getRemoveHostMapCmd(volumeName string, hostId string) string {
	return fmt.Sprintf("rmvdiskhostmap -host %s %s", hostId, volumeName)
}

/*
# 拼装 Server multipath WWID
wwid = '3' + vdisk_UID.lower()*/
func (fc *FcClient) getvDiskUIDByVolume(volumeName string) (vdiskUid string, err error) {
	out, errInfo, err := fc.runSsh("lsvdisk -delim '|'")
	if err != nil || errInfo != "" {
		return "", commonErrFmtNLog(fmt.Sprintf("lsvdisk  %s", volumeName), out, errInfo, err)
	}
	log.Infof("lsvdisk %s succeed out: [%s], err [%s]", volumeName, out, errInfo)

	vdisUIDIndex := getVdiskUIDIndex(out)
	if vdisUIDIndex == 0 {
		return "", commonErrFmtNLog(fmt.Sprintf("lsvdisk  %s", volumeName), out, errInfo, err)
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if len(strings.Split(line, "|")) > 1 && strings.Split(line, "|")[1] == volumeName {
			return strings.Split(line, "|")[vdisUIDIndex], nil
		}
	}
	return "", fmt.Errorf("lsvdisk volumeName absent")
}

func (fc *FcClient) GetWWIdByVolume(volumeName string) (string, error) {
	vdiskUid, err := fc.getvDiskUIDByVolume(volumeName)
	if err != nil {
		return "", err
	}
	wwid := "3" + strings.ToLower(vdiskUid)
	return wwid, nil
}

func (fc *FcClient) RemoveAllMapping(volumeName string) error {
	out, errInfo, err := fc.runSsh(fmt.Sprintf("lsvdiskhostmap %s", volumeName), volumeName, "removeAllMapping")
	if err != nil || errInfo != "" {
		if !strings.Contains(errInfo, "does not meet the naming rules") &&
			!strings.Contains(errInfo, "does not meet the naming rules") {
			return commonErrFmtNLog(fmt.Sprintf("lsvdiskhostmap [%s]", volumeName), out, errInfo, err)
		}
		log.Infof("lsvdiskhostmap [%s] is empty, return success: [%s], err: [%v], errInfo: [%s]", volumeName, out, err, errInfo)
		return nil
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, volumeName) {
			reg := regexp.MustCompile(`\s+`)
			columes := reg.Split(strings.TrimSpace(line), -1)
			// rmvdiskhostmap -host ${host_id} ${vdisk_name}
			if len(columes) <= 3 {
				return fmt.Errorf("lsvdiskhostmap [%s] get failed: %s,%s", volumeName, columes, line)
			}
			out, errInfo, err := fc.runSsh(getRemoveHostMapCmd(volumeName, columes[3]), volumeName, "removeAllMapping")
			if err != nil || errInfo != "" {
				return commonErrFmtNLog(fmt.Sprintf("rmvdiskhostmap [%s]", volumeName), out, errInfo, err)
			}
			log.Infof("rmvdiskhostmap [%s] succeed out: [%s]", volumeName, out)
		}
	}
	return nil
}

func (fc *FcClient) DeleteVolume(volumeName string) error {
	out, errInfo, err := fc.runSsh(fmt.Sprintf("rmvdisk %s", volumeName), volumeName, "deleteVolume")
	if (err != nil || errInfo != "") && !strings.Contains(errInfo, "not exist or is not a suitable candidate") {
		return commonErrFmtNLog(fmt.Sprintf("deleteVolume [%s]", volumeName), out, errInfo, err)
	}
	log.Infof("deleteVolume [%s] succeed out: [%s]", volumeName, out)
	return nil
}

type DiskInfo struct {
	HostName string
	ScsiId   string
	Wwid     string
}

func regexLsHostVDiskMap(out, volumeName string) (diskInfos []DiskInfo) {
	//2  h_192.168.1.205 10      37       ppaschujie04-1577         60050767088080A26800        00000000025C 0           io_grp0       private                                        \n], err []"
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		reg := regexp.MustCompile(`\s+`)
		columns := reg.Split(strings.TrimSpace(line), -1)
		if strings.Contains(line, volumeName) {
			if len(columns) <= 3 {
				continue
			}
		}
		if len(columns) <= 6 {
			continue
		}
		diskInfos = append(diskInfos, DiskInfo{
			HostName: columns[1],
			ScsiId:   columns[2],
			Wwid:     "3" + strings.ToLower(columns[5]),
		})
	}
	log.Infof("regexLsHostVDiskMap diskInfo: %+v", diskInfos)
	return
}

type DeviceBlock struct {
	//h is the HBA number,c is the channel on the HBA,t is the SCSI target ID,and l is the LUN.
	hctl        string
	blockDevice string
	wwid        string
}

func getAbnormalDeviceByScsiIdAndWwid(hostIP, volumeName, scsiId, wwid string) (error, []*DeviceBlock) {
	var deviceBlock []*DeviceBlock

	getScsiCmd := fmt.Sprintf("lsscsi -i |grep -E 'INSPUR'|grep :%s]|grep -v %s |awk '{print $1 $6 $7}'", scsiId, wwid)

	out, errInfo, err := RunSSHLocalCMD(getScsiCmd, hostIP, hostIP, volumeName, "getAbnormalDeviceByScsiIdAndWwid")
	if err != nil {
		log.Errorf("getAbnormalDeviceByScsiIdAndWwid failed out (wwid: %s): [%s], err [%s] [%v]", wwid, volumeName, errInfo, string(out))
		return fmt.Errorf("lsscsi -i failed"), deviceBlock
	}
	log.Infof("getAbnormalDeviceByScsiIdAndWwid succeed out (wwid: %s): [%v], dev [%s]", wwid, volumeName, string(out))

	devices := strings.Split(out, "\n")
	for _, device := range devices {
		if len(device) > 33 {
			log.Infof("getAbnormalDeviceByScsiIdAndWwid DeviceBlock spilt (wwid: %s): [%v], dev [%s]", wwid, volumeName, device)
			blockBIdx := strings.Index(device, "]") + 1
			blockEIdx := len(device) - 33
			deviceBlock = append(deviceBlock, &DeviceBlock{
				//[16:0:12:1] => 16:0:12:1
				hctl: device[1 : blockBIdx-1],
				///dev/sdao => sdao
				blockDevice: device[blockBIdx:blockEIdx],
				wwid:        device[blockEIdx:],
			})
		}
	}
	return nil, deviceBlock
}

func (fc *FcClient) RemoveAbnormalDevice(nodesIP map[string]string, volumeName, wwid string) error {

	/*get lshostvdiskmap*/
	lsHostVDistMapCmd := fmt.Sprintf("lshostvdiskmap | grep %s", volumeName)
	out, errInfo, err := fc.runSsh(lsHostVDistMapCmd, volumeName, "removeAbnormalDevice")
	if err != nil || errInfo != "" {
		log.Errorf("removeAbnormalDevice lshostvdiskmap volume: %s wwid: %s failed , err [%s]", volumeName, wwid, errInfo)
		return errors.New(fmt.Sprintf("lshostvdiskmap failed%s,%s", out, errInfo))

	} else if !strings.Contains(out, volumeName) {
		return errors.New(fmt.Sprintf("lshostvdiskmap out: %s, not found volumeName: %s", out, volumeName))
	}

	log.Infof("removeAbnormalDevice lshostvdiskmap volume: %s wwid: %s succeed , err [%s]", volumeName, wwid, errInfo)
	diskInfos := regexLsHostVDiskMap(out, volumeName)
	if len(diskInfos) == 0 {
		log.Infof("removeAbnormalDevice is nil volume: %s wwid: %s", volumeName, wwid)
		return nil
	}

	for _, diskInfo := range diskInfos {

		nodeIP := ""
		ok := false
		if nodeIP, ok = nodesIP[diskInfo.HostName]; !ok {
			return fmt.Errorf("k8s node not found name: %s", diskInfo.HostName)
		}
		log.V(5).Infof("getAbnormalDeviceByScsiIdAndWwid in node: %s, ip: %s", diskInfo.HostName, nodeIP)
		err, abnormalDevices := getAbnormalDeviceByScsiIdAndWwid(nodeIP, volumeName, diskInfo.ScsiId, diskInfo.Wwid)

		for _, abnormalDevice := range abnormalDevices {

			delDevMapCmd := fmt.Sprintf("if [ -f /sys/class/scsi_device/%s/device/delete ]; then echo 1 > /sys/class/scsi_device/%s/device/delete; fi", abnormalDevice.hctl, abnormalDevice.hctl)

			if out, errInfo, err = RunSSHLocalCMD(delDevMapCmd, nodeIP, nodeIP, volumeName, "removeAbnormalDevice"); err != nil || errInfo != "" {
				log.Errorf("removeAbnormalDevice failed out: (source wwid: %s|target wwid: %s)[%s], h.c.t.l [%s], blockdev[%s], err [%s]",
					abnormalDevice.wwid, wwid, volumeName, abnormalDevice.hctl, abnormalDevice.blockDevice, errInfo)
			}
			log.Infof("removeAbnormalDevice success:(source wwid: %s|target wwid: %s)[%s], h.c.t.l, blockdev[%s], [%s]", abnormalDevice.wwid, wwid, volumeName, abnormalDevice.blockDevice, abnormalDevice.hctl)
		}
	}

	return nil
}
