package types

type FileData struct {
	PartialDataInfo *PartialDataInfo
	ContentType     string
	ContentSize     int64
	Content         []byte
}

type PartialDataInfo struct {
	RangeStartByte int64
	TotalDataSize  int64
}
