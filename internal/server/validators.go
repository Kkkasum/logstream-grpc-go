package server

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "logstream/pkg/api/logstream"
)

func validateSaveLogRequest(req *pb.SaveLogRequest) error {
	var violations []*errdetails.BadRequest_FieldViolation

	log := req.GetLog()
	if log == nil {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log",
			Description: "empty",
		})

		st, err := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).
			WithDetails(&errdetails.BadRequest{
				FieldViolations: violations,
			})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		return st.Err()
	}

	if source := log.GetSource(); len(source) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.source",
			Description: "empty",
		})
	}

	if level := log.GetLevel(); level > 2 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.level",
			Description: "invalid value",
		})
	}

	if message := log.GetMessage(); len(message) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.message",
			Description: "empty",
		})
	}

	if timestamp := log.GetTimestamp(); timestamp == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.timestamp",
			Description: "empty",
		})
	}

	if len(violations) > 0 {
		st, err := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).
			WithDetails(&errdetails.BadRequest{
				FieldViolations: violations,
			})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		return st.Err()
	}

	return nil
}

func validateListLogRequest(req *pb.ListLogRequest) error {
	var violations []*errdetails.BadRequest_FieldViolation

	if id := req.GetId(); id == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "id",
			Description: "empty",
		})
	}

	if len(violations) > 0 {
		st, err := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).
			WithDetails(&errdetails.BadRequest{
				FieldViolations: violations,
			})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		return st.Err()
	}

	return nil
}

func validateListLogsRequest(req *pb.ListLogsRequest) error {
	var violations []*errdetails.BadRequest_FieldViolation

	if level := req.GetLevel(); level > 2 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.level",
			Description: "invalid value",
		})
	}

	if startTime := req.GetStartTime(); startTime == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "start_time",
			Description: "empty",
		})
	}

	if endTime := req.GetStartTime(); endTime == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "end_time",
			Description: "empty",
		})
	}

	if len(violations) > 0 {
		st, err := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).
			WithDetails(&errdetails.BadRequest{
				FieldViolations: violations,
			})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		return st.Err()
	}

	return nil
}

func validateListLogsStreamRequest(req *pb.ListLogsStreamRequest) error {
	var violations []*errdetails.BadRequest_FieldViolation

	if level := req.GetLevel(); level > 2 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.level",
			Description: "invalid value",
		})
	}

	if startTime := req.GetStartTime(); startTime == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "start_time",
			Description: "empty",
		})
	}

	if endTime := req.GetStartTime(); endTime == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "end_time",
			Description: "empty",
		})
	}

	if len(violations) > 0 {
		st, err := status.New(codes.InvalidArgument, codes.InvalidArgument.String()).
			WithDetails(&errdetails.BadRequest{
				FieldViolations: violations,
			})
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		return st.Err()
	}

	return nil
}
