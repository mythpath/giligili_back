package replica

import (
	"io"
	"fmt"
	"encoding/binary"
	"encoding/json"
)

type EventType uint8

const (
	UNKNOWN_EVENT 			 = 0
	START_EVENT_V3           = 1
	QUERY_EVENT              = 2
	STOP_EVENT               = 3
	ROTATE_EVENT             = 4
	INTVAR_EVENT             = 5
	LOAD_EVENT               = 6
	SLAVE_EVENT              = 7
	CREATE_FILE_EVENT        = 8
	APPEND_BLOCK_EVENT       = 9
	EXEC_LOAD_EVENT          = 10
	DELETE_FILE_EVENT        = 11
	NEW_LOAD_EVENT           = 12
	RAND_EVENT               = 13
	USER_VAR_EVENT           = 14
	FORMAT_DESCRIPTION_EVENT = 15
	XID_EVENT                = 16
	BEGIN_LOAD_QUERY_EVENT   = 17
	EXECUTE_LOAD_QUERY_EVENT = 18
	TABLE_MAP_EVENT          = 19
	WRITE_ROWS_EVENT_V0	     = 20
	UPDATE_ROWS_EVENT_V0 	 = 21
	DELETE_ROWS_EVENT_V0 	 = 22
	WRITE_ROWS_EVENT_V1      = 23
	UPDATE_ROWS_EVENT_V1     = 24
	DELETE_ROWS_EVENT_V1     = 25
	INCIDENT_EVENT           = 26
	HEARTBEAT_EVENT      	 = 27
	IGNORABLE_EVENT      	 = 28
	ROWS_QUERY_EVENT     	 = 29
	WRITE_ROWS_EVENT_V2      = 30
	UPDATE_ROWS_EVENT_V2     = 31
	DELETE_ROWS_EVENT_V2     = 32
	GTID_EVENT           	 = 33
	ANONYMOUS_GTID_EVENT 	 = 34
	PREVIOUS_GTIDS_EVENT 	 = 35
)

func (e EventType) String() string {
	switch e {
	case UNKNOWN_EVENT:
		return "UnknownEvent"
	case START_EVENT_V3:
		return "StartEventV3"
	case QUERY_EVENT:
		return "QueryEvent"
	case STOP_EVENT:
		return "StopEvent"
	case ROTATE_EVENT:
		return "RotateEvent"
	case INTVAR_EVENT:
		return "IntVarEvent"
	case LOAD_EVENT:
		return "LoadEvent"
	case SLAVE_EVENT:
		return "SlaveEvent"
	case CREATE_FILE_EVENT:
		return "CreateFileEvent"
	case APPEND_BLOCK_EVENT:
		return "AppendBlockEvent"
	case EXEC_LOAD_EVENT:
		return "ExecLoadEvent"
	case DELETE_FILE_EVENT:
		return "DeleteFileEvent"
	case NEW_LOAD_EVENT:
		return "NewLoadEvent"
	case RAND_EVENT:
		return "RandEvent"
	case USER_VAR_EVENT:
		return "UserVarEvent"
	case FORMAT_DESCRIPTION_EVENT:
		return "FormatDescriptionEvent"
	case XID_EVENT:
		return "XIDEvent"
	case BEGIN_LOAD_QUERY_EVENT:
		return "BeginLoadQueryEvent"
	case EXECUTE_LOAD_QUERY_EVENT:
		return "ExectueLoadQueryEvent"
	case TABLE_MAP_EVENT:
		return "TableMapEvent"
	case WRITE_ROWS_EVENT_V0:
		return "WriteRowsEventV0"
	case UPDATE_ROWS_EVENT_V0:
		return "UpdateRowsEventV0"
	case DELETE_ROWS_EVENT_V0:
		return "DeleteRowsEventV0"
	case WRITE_ROWS_EVENT_V1:
		return "WriteRowsEventV1"
	case UPDATE_ROWS_EVENT_V1:
		return "UpdateRowsEventV1"
	case DELETE_ROWS_EVENT_V1:
		return "DeleteRowsEventV1"
	case INCIDENT_EVENT:
		return "IncidentEvent"
	case HEARTBEAT_EVENT:
		return "HeartbeatEvent"
	case IGNORABLE_EVENT:
		return "IgnorableEvent"
	case ROWS_QUERY_EVENT:
		return "RowsQueryEvent"
	case WRITE_ROWS_EVENT_V2:
		return "WriteRowsEventV2"
	case UPDATE_ROWS_EVENT_V2:
		return "UpdateRowsEventV2"
	case DELETE_ROWS_EVENT_V2:
		return "DeleteRowsEventV2"
	case GTID_EVENT:
		return "GTIDEvent"
	case ANONYMOUS_GTID_EVENT:
		return "AnonymousGTIDEvent"
	case PREVIOUS_GTIDS_EVENT:
		return "PreviousGTIDsEvent"

	default:
		return "UnknownEvent"
	}
}

const (
	MagicHeaderSize 	= 4
	EventHeaderSize 	= 19
)

var (
	//binlog header [ fe `bin` ]
	BinLogFileMagicNumber []byte = []byte{0xfe, 0x62, 0x69, 0x6e}
)

// binlog public header
const (
	EventHeaderZero	 		= 0
	EventHeaderTimestamp	= 4
	EventHeaderEventType	= 5
	EventHeaderServerId		= 9
	EventHeaderEventLength	= 13
	EventHeaderNextPosition	= 17
	EventHeaderFlags		= 19
)

const (
	PreviousBodyUUIdCount	= 8
	PreviousBodyUUId 		= 16
	PreviousBodyTId			= 8
)

type BinlogEvent struct {
	Header 		*EventHeader
	RawData 	[]byte
}

type Event interface {
	Decode([]byte) error

	Print(io.Writer)
}

type EventHeader struct {
	// 0 : 4
	Timestamp    	uint32
	// 4 : 1
	EventType    	uint8
	// 5 : 4
	ServerId     	uint32
	// 9 : 4
	EventSize  		uint32
	// 13 : 4
	NextLogPosition uint32
	// 17 : 2
	Flags        	uint16
}

func (e *EventHeader) Decode(data []byte) error {
	if len(data) < EventHeaderSize {
		return fmt.Errorf("event header size is too short, expect: %d, got: %d", EventHeaderSize, len(data))
	}

	offset := 0
	e.Timestamp = binary.LittleEndian.Uint32(data[offset:])

	offset += 4
	e.EventType = data[offset]

	offset += 1
	e.ServerId = binary.LittleEndian.Uint32(data[offset:])

	offset += 4
	e.EventSize = binary.LittleEndian.Uint32(data[offset:])
	if e.EventSize < EventHeaderSize {
		return fmt.Errorf("invalid event size, it must be greater `%d`", EventHeaderSize)
	}

	offset += 4
	e.NextLogPosition = binary.LittleEndian.Uint32(data[offset:])

	offset += 4
	e.Flags = binary.LittleEndian.Uint16(data[offset:])

	return nil
}

// json encoder
func (e *EventHeader) Print(w io.Writer) {
	json.NewEncoder(w).Encode(e)
}
