package v1

// from new to old
type BackupSorterByCreateTime []RedisBackup

func (b BackupSorterByCreateTime) Len() int {
	return len(b)
}

func (b BackupSorterByCreateTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b BackupSorterByCreateTime) Less(i, j int) bool {
	return b[i].CreationTimestamp.After(b[j].CreationTimestamp.Time)
}
