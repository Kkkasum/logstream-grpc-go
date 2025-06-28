package repo

import pb "logstream/pkg/api/logstream"

type Log struct {
	Id        *int32 `db:"id"`
	Source    string `db:"source"`
	Level     int32  `db:"lvl"`
	Message   string `db:"message"`
	CreatedAt int64  `db:"created_at"`
}

func FromPbLog(l *pb.Log) *Log {
	return &Log{
		Id:        l.Id,
		Source:    l.Source,
		Level:     int32(l.Level),
		Message:   l.Message,
		CreatedAt: l.Timestamp,
	}
}

func (l *Log) ToPbLog() *pb.Log {
	return &pb.Log{
		Id:        l.Id,
		Source:    l.Source,
		Level:     pb.Level(l.Level),
		Message:   l.Message,
		Timestamp: l.CreatedAt,
	}
}
