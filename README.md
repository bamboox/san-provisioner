# san-provisioner

this san-provisioner support 

- IBM FC SAN 
- ALIYUN INSPUR

[IBM Flash Storage](https://www.ibm.com/it-infrastructure/storage/flash)

# Feature

1. annotation

- volume.beta.kubernetes.io/thin-rate # default none
- volume.beta.kubernetes.io/mount-options # default none
- volume.beta.kubernetes.io/mkfs-fstype # default ext4

# Quick Start

1. install 

kubctl apply -f ./deploy 

2. create test pvc

kubctl apply -f ./test/test-pvc.yaml

2. create test pod

kubctl apply -f ./test/test-pod.yaml


# ISSUES

1. FailedMount
erro info
```
Warning  FailedMount             11s (x6 over 50s)  kubelet, dbm02           MountVolume.WaitForAttach failed for volume "pvc-53918fcd-c350-11e9-8c87-50af732f4b85" : no fc disk found
```

```
ssh superuser@192.168.1.2 lshostvdiskmap | grep -i 845 #
ls /dev/disk/by-id/ | grep 845 # not found
```
fix this 
delete by SCSI ID
```shell
scsi_id=${scsi_id}
wwid=${wwid}
devices=lsscsi -i |grep MCS | grep ":${scsi_id}]" | grep -v $wwid | awk '{print $1}'
for i in $devices; do device=${i:1:-1} ; echo 1 >/sys/class/scsi_device/$device/device/delete ; done
```
