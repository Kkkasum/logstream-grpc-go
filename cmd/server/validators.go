package main

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

	source := log.GetSource()
	if len(source) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.source",
			Description: "empty",
		})
	}

	message := log.GetMessage()
	if len(message) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "log.message",
			Description: "empty",
		})
	}

	timestamp := log.GetTimestamp()
	if timestamp == 0 {
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

func validateListLogsRequest(req *pb.ListLogsRequest) error {
	var violations []*errdetails.BadRequest_FieldViolation

	startTime := req.GetStartTime()
	if startTime == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "start_time",
			Description: "empty",
		})
	}

	endTime := req.GetStartTime()
	if endTime == 0 {
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
