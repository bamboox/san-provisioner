## san 挂载问题排查	
1. FailedMount

Warning  FailedMount             11s (x6 over 50s)  kubelet, dbm02           MountVolume.WaitForAttach failed for volume "pvc-53918fcd-c350-11e9-8c87-50af732f4b85" : no fc disk found


ls /dev/disk/by-id/ | grep 845 #

ssh superuser@192.168.1.2 lshostvdiskmap | grep -i 845

## delete by SCSI ID
scsi_id=${scsi_id}
wwid=${wwid}
devices=lsscsi -i |grep MCS | grep ":${scsi_id}]" | grep -v $wwid | awk '{print $1}'
for i in $devices; do device=${i:1:-1} ; echo 1 >/sys/class/scsi_device/$device/device/delete ; done


## format
挂载如果是非 readonly 需要注意format问题，数据会丢失

# Annotation

- ThinRateAnnotation    = "volume.beta.kubernetes.io/thin-rate"

- MountOptionAnnotation = "volume.beta.kubernetes.io/mount-options"
 
- MkfsFsTypeAnnotation  = "volume.beta.kubernetes.io/mkfs-fstype"  
  default ext4