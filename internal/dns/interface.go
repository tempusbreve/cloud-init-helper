package dns

import "context"

type API interface {
	GetRecords(ctx context.Context, recordName, recordType string) ([]Record, error)
	CreateMXRecord(ctx context.Context, name string, mxHost string, weight int) error
	GetRecord(ctx context.Context, rid any) (Record, error)
	CreateRecord(ctx context.Context, record Record) error
	UpdateRecord(ctx context.Context, record Record) error
	DeleteRecord(ctx context.Context, rid any) error
}

type RecordType string

func (r RecordType) String() string { return string(r) }

const (
	RecordTypeA     = RecordType("A")
	RecordTypeAAAA  = RecordType("AAAA")
	RecordTypeCNAME = RecordType("CNAME")
	RecordTypeMX    = RecordType("MX")
	RecordTypeNS    = RecordType("NS")
	RecordTypeSRV   = RecordType("SRV")
	RecordTypeTXT   = RecordType("TXT")
)

type Record interface {
	ID() any
	Name() string
	Type() RecordType
	Content() string
	TTL() int
}

func NewRecord(name string, rtype RecordType, content string) Record {
	return record{
		name:    name,
		rtype:   string(rtype),
		content: content,
	}
}

type record struct {
	id      any
	name    string
	rtype   string
	content string
	ttl     int
}

func (r record) ID() any          { return r.id }
func (r record) Name() string     { return r.name }
func (r record) Type() RecordType { return RecordType(r.rtype) }
func (r record) Content() string  { return r.content }
func (r record) TTL() int         { return r.ttl }
