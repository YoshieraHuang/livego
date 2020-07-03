package h264

import (
	"bytes"
	"fmt"
	"io"
)

const (
	iFrame byte = 0
	pFrame byte = 1
	bFrame byte = 2
)

const (
	naluTypeNotDefine byte = 0
	//slice_layer_without_partioning_rbsp() sliceheader
	naluTypeSlice byte = 1
	// slice_data_partition_a_layer_rbsp( ), slice_header
	naluTypeDpa byte = 2
	// slice_data_partition_b_layer_rbsp( )
	naluTypeDpb byte = 3
	// slice_data_partition_c_layer_rbsp( )
	naluTypeDpc byte = 4
	// slice_layer_without_partitioning_rbsp( ),sliceheader
	naluTypeIdr byte = 5
	//sei_rbsp( )
	naluTypeSei byte = 6
	//seq_parameter_set_rbsp( )
	naluTypeSps byte = 7
	//pic_parameter_set_rbsp( )
	naluTypePps byte = 8
	// access_unit_delimiter_rbsp( )
	naluTypeAud byte = 9
	//end_of_seq_rbsp( )
	naluTypeEoesq byte = 10
	//end_of_stream_rbsp( )
	naluTypeEostream byte = 11
	//filler_data_rbsp( )
	naluTypeFiller byte = 12
)

const (
	naluBytesLen int = 4
	maxSpsPpsLen int = 2 * 1024
)

var (
	// ErrDecDataNil means dec buf is nil
	ErrDecDataNil = fmt.Errorf("dec buf is nil")
	// ErrSpsData means sps data error
	ErrSpsData = fmt.Errorf("sps data error")
	// ErrPpsHeader means pps header error
	ErrPpsHeader = fmt.Errorf("pps header error")
	// ErrPpsData means pps data error
	ErrPpsData = fmt.Errorf("pps data error")
	// ErrInvalidNaluHeader means invalid nalu header
	ErrInvalidNaluHeader = fmt.Errorf("invalid nalu header")
	// ErrInvalidVideoData means invalid video data
	ErrInvalidVideoData = fmt.Errorf("invalid video data")
	// ErrDataSizeNotMatch means data size not match
	ErrDataSizeNotMatch = fmt.Errorf("data size not match")
	// ErrNaluBodyLen means nalu body len error
	ErrNaluBodyLen = fmt.Errorf("nalu body len error")
)

var startCode = []byte{0x00, 0x00, 0x00, 0x01}
var naluAud = []byte{0x00, 0x00, 0x00, 0x01, 0x09, 0xf0}

// Parser is a h.264 parser
type Parser struct {
	frameType    byte
	specificInfo []byte
	pps          *bytes.Buffer
}

type sequenceHeader struct {
	configVersion        byte //8bits
	avcProfileIndication byte //8bits
	profileCompatility   byte //8bits
	avcLevelIndication   byte //8bits
	reserved1            byte //6bits
	naluLen              byte //2bits
	reserved2            byte //3bits
	spsNum               byte //5bits
	ppsNum               byte //8bits
	spsLen               int
	ppsLen               int
}

// NewParser returns a parser
func NewParser() *Parser {
	return &Parser{
		pps: bytes.NewBuffer(make([]byte, maxSpsPpsLen)),
	}
}

//return value 1:sps, value2 :pps
func (parser *Parser) parseSpecificInfo(src []byte) error {
	if len(src) < 9 {
		return ErrDecDataNil
	}
	sps := []byte{}
	pps := []byte{}

	var seq sequenceHeader
	seq.configVersion = src[0]
	seq.avcProfileIndication = src[1]
	seq.profileCompatility = src[2]
	seq.avcLevelIndication = src[3]
	seq.reserved1 = src[4] & 0xfc
	seq.naluLen = src[4]&0x03 + 1
	seq.reserved2 = src[5] >> 5

	//get sps
	seq.spsNum = src[5] & 0x1f
	seq.spsLen = int(src[6])<<8 | int(src[7])

	if len(src[8:]) < seq.spsLen || seq.spsLen <= 0 {
		return ErrSpsData
	}
	sps = append(sps, startCode...)
	sps = append(sps, src[8:(8+seq.spsLen)]...)

	//get pps
	tmpBuf := src[(8 + seq.spsLen):]
	if len(tmpBuf) < 4 {
		return ErrPpsHeader
	}
	seq.ppsNum = tmpBuf[0]
	seq.ppsLen = int(0)<<16 | int(tmpBuf[1])<<8 | int(tmpBuf[2])
	if len(tmpBuf[3:]) < seq.ppsLen || seq.ppsLen <= 0 {
		return ErrPpsData
	}

	pps = append(pps, startCode...)
	pps = append(pps, tmpBuf[3:]...)

	parser.specificInfo = append(parser.specificInfo, sps...)
	parser.specificInfo = append(parser.specificInfo, pps...)

	return nil
}

func (parser *Parser) isNaluHeader(src []byte) bool {
	if len(src) < naluBytesLen {
		return false
	}
	return src[0] == 0x00 &&
		src[1] == 0x00 &&
		src[2] == 0x00 &&
		src[3] == 0x01
}

func (parser *Parser) naluSize(src []byte) (int, error) {
	if len(src) < naluBytesLen {
		return 0, fmt.Errorf("nalusizedata invalid")
	}
	buf := src[:naluBytesLen]
	size := int(0)
	for i := 0; i < len(buf); i++ {
		size = size<<8 + int(buf[i])
	}
	return size, nil
}

func (parser *Parser) getAnnexbH264(src []byte, w io.Writer) error {
	dataSize := len(src)
	if dataSize < naluBytesLen {
		return ErrInvalidVideoData
	}
	parser.pps.Reset()
	_, err := w.Write(naluAud)
	if err != nil {
		return err
	}

	index := 0
	nalLen := 0
	hasSpsPps := false
	hasWriteSpsPps := false

	for dataSize > 0 {
		nalLen, err = parser.naluSize(src[index:])
		if err != nil {
			return ErrDataSizeNotMatch
		}
		index += naluBytesLen
		dataSize -= naluBytesLen
		if dataSize >= nalLen && len(src[index:]) >= nalLen && nalLen > 0 {
			nalType := src[index] & 0x1f
			switch nalType {
			case naluTypeAud:
			case naluTypeIdr:
				if !hasWriteSpsPps {
					hasWriteSpsPps = true
					if !hasSpsPps {
						if _, err := w.Write(parser.specificInfo); err != nil {
							return err
						}
					} else {
						if _, err := w.Write(parser.pps.Bytes()); err != nil {
							return err
						}
					}
				}
				fallthrough
			case naluTypeSlice:
				fallthrough
			case naluTypeSei:
				_, err := w.Write(startCode)
				if err != nil {
					return err
				}
				_, err = w.Write(src[index : index+nalLen])
				if err != nil {
					return err
				}
			case naluTypeSps:
				fallthrough
			case naluTypePps:
				hasSpsPps = true
				_, err := parser.pps.Write(startCode)
				if err != nil {
					return err
				}
				_, err = parser.pps.Write(src[index : index+nalLen])
				if err != nil {
					return err
				}
			}
			index += nalLen
			dataSize -= nalLen
		} else {
			return ErrNaluBodyLen
		}
	}
	return nil
}

// Parse parses the data
func (parser *Parser) Parse(b []byte, isSeq bool, w io.Writer) (err error) {
	switch isSeq {
	case true:
		err = parser.parseSpecificInfo(b)
	case false:
		// is annexb
		if parser.isNaluHeader(b) {
			_, err = w.Write(b)
		} else {
			err = parser.getAnnexbH264(b, w)
		}
	}
	return
}
