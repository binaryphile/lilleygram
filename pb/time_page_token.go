package pb

import (
	"encoding/base64"
	"google.golang.org/protobuf/proto"
)

func NewTimePageToken(updated_at int64, id uint64, pageSize uint32) *TimePageToken {
	return &TimePageToken{
		Id:        id,
		PageSize:  pageSize,
		UpdatedAt: updated_at,
	}
}

func (t *TimePageToken) Marshal() string {
	token, err := proto.Marshal(t)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(token)
}

func PageSize(t *TimePageToken) uint32 {
	return t.PageSize
}

func UnmarshalTimePageToken(pageToken string) *TimePageToken {
	if pageToken == "" {
		return nil
	}

	byteToken, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		return nil
	}

	token := &TimePageToken{}

	err = proto.Unmarshal(byteToken, token)
	if err != nil {
		return nil
	}

	return token
}
