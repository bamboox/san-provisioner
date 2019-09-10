module github.com/kubernetes-sigs/san-client

go 1.12

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.26.0
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.2.1
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.2
	go.uber.org/atomic => go.uber.org/atomic v0.0.0-20181018215023-8dc6146f7569
	go.uber.org/multierr => go.uber.org/multierr v0.0.0-20180122172545-ddea229ff1df
	go.uber.org/zap => go.uber.org/zap v0.0.0-20180814183419-67bc79d13d15
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/exp => github.com/golang/exp v0.0.0-20190312203227-4b39c73a6495
	golang.org/x/image => github.com/golang/image v0.0.0-20190802002840-cff245a6509b
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190313153728-d0100b6bd8b3
	golang.org/x/mobile => github.com/golang/mobile v0.0.0-20190814143026-e8b3e6111d02
	golang.org/x/net => github.com/golang/net v0.0.0-20190813141303-74dc4d7220e7
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sync => github.com/golang/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys => github.com/golang/sys v0.0.0-20190813064441-fde4db37ae7a
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/time => github.com/golang/time v0.0.0-20190308202827-9d24e82272b4
	golang.org/x/tools => github.com/golang/tools v0.0.0-20190311212946-11955173bddd
	google.golang.org/api => github.com/googleapis/google-api-go-client v0.0.0-20181220000619-583d854617af
	google.golang.org/appengine => github.com/golang/appengine v1.5.0
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20180817151627-c66870c02cf8
	google.golang.org/grpc => github.com/grpc/grpc-go v1.23.0
	k8s.io/api => github.com/kubernetes/api v0.0.0-20190819141256-463df2a5c347
	k8s.io/apiextensions-apiserver => github.com/kubernetes/apiextensions-apiserver v0.0.0-20190819143642-035c9555f1df
	k8s.io/apimachinery => github.com/kubernetes/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver => github.com/kubernetes/apiserver v0.0.0-20190819142451-3e05a936e664
	k8s.io/cli-runtime => github.com/kubernetes/cli-runtime v0.0.0-20190819144026-806ed77012ad
	k8s.io/client-go => github.com/kubernetes/client-go v0.0.0-20190819141728-80a1a93c2f21
	k8s.io/cloud-provider => github.com/kubernetes/cloud-provider v0.0.0-20190819145150-e4132ada86f0
	k8s.io/cluster-bootstrap => github.com/kubernetes/cluster-bootstrap v0.0.0-20190819145007-4a236b5010c9
	k8s.io/code-generator => github.com/kubernetes/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => github.com/kubernetes/component-base v0.0.0-20190819141909-7554603fbbcc
	k8s.io/cri-api => github.com/kubernetes/cri-api v0.0.0-20190817025403-3ae76f584e79
	k8s.io/csi-translation-lib => github.com/kubernetes/csi-translation-lib v0.0.0-20190819145326-779f25e39d41
	k8s.io/gengo => github.com/kubernetes/gengo v0.0.0-20190116091435-f8a0810f38af
	k8s.io/heapster => github.com/kubernetes/heapster v1.2.0-beta.1
	k8s.io/klog => github.com/kubernetes/klog v0.4.0
	k8s.io/kube-aggregator => github.com/kubernetes/kube-aggregator v0.0.0-20190819142801-b7d64bc74d80
	k8s.io/kube-controller-manager => github.com/kubernetes/kube-controller-manager v0.0.0-20190819144833-b3c478023999
	k8s.io/kube-openapi => github.com/kubernetes/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kube-proxy => github.com/kubernetes/kube-proxy v0.0.0-20190819144346-b5160da64689
	k8s.io/kube-scheduler => github.com/kubernetes/kube-scheduler v0.0.0-20190819144658-be4fac0251cb
	k8s.io/kubelet => github.com/kubernetes/kubelet v0.0.0-20190819144523-3c1be3fbd485
	k8s.io/kubernetes => github.com/kubernetes/kubernetes v1.15.3
	k8s.io/legacy-cloud-providers => github.com/kubernetes/legacy-cloud-providers v0.0.0-20190819145512-fe39bd0ea42f
	k8s.io/metrics => github.com/kubernetes/metrics v0.0.0-20190819143843-3ba45e11778e
	k8s.io/repo-infra => github.com/kubernetes/repo-infra v0.0.0-20181204233714-00fe14e3d1a3
	k8s.io/sample-apiserver => github.com/kubernetes/sample-apiserver v0.0.0-20190819143050-b0bd735e1fff
	k8s.io/utils => github.com/kubernetes/utils v0.0.0-20190221042446-c2654d5206da
	sigs.k8s.io/kustomize => github.com/kubernetes-sigs/kustomize v2.0.3+incompatible
	sigs.k8s.io/sig-storage-lib-external-provisioner => github.com/kubernetes-sigs/sig-storage-lib-external-provisioner v4.0.0+incompatible
	sigs.k8s.io/structured-merge-diff => github.com/kubernetes-sigs/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
	sigs.k8s.io/yaml => github.com/kubernetes-sigs/yaml v1.1.0
)

require (
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/time v0.0.0-20161028155119-f51c12702a4d
	google.golang.org/grpc v1.13.0
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/cloud-provider v0.0.0
	k8s.io/klog v0.3.1
	k8s.io/kubernetes v1.15.3
	sigs.k8s.io/sig-storage-lib-external-provisioner v4.0.0+incompatible
)
