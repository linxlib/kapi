package internal

import "strings"

func GetCommentAfterPrefix(comment string, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(comment, prefix))
}
