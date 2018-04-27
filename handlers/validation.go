package handlers

import (
	"strconv"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// ValidateAccountID - performs validation of the accountID
// from the URL parameter, returns the parsed integer account
// id as well as an error if there were problems validating
func ValidateAccountID(ctx echo.Context) (int32, error) {
	var (
		aid int64
		err error
	)
	if aidStr := ctx.Param("account"); aidStr != "" {
		if aid, err = strconv.ParseInt(aidStr, 10, 32); err != nil || aid < 1 {
			return 0, errors.Wrap(err, "failed to convert account id url param to integer")
		}
		return int32(aid), nil
	}
	return 0, errors.New("accountID invalid, missing or empty")
}

// ValidateCheckUUID - performs validation of the accountID
// from the URL parameter, returns the parsed integer account
// id as well as an error if there were problems validating
func ValidateCheckUUID(ctx echo.Context) (uuid.UUID, error) {
	var (
		u   uuid.UUID
		err error
	)
	if u, err = uuid.FromString(ctx.Param("check_uuid")); err == nil {
		return u, nil
	}
	return u, errors.New("check_uuid invalid, missing or empty")
}

// ValidateCheckName - performs validation of the check_name
// from the URL parameter, returns the parsed string check_name
// as well as an error if there were problems validating
func ValidateCheckName(ctx echo.Context) (string, error) {
	var (
		name string
	)
	if name = ctx.Param("check_name"); name != "" {
		return name, nil
	}
	return name, errors.New("check_name invalid, missing or empty")
}
