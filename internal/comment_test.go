package internal

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestGetCommentAfterPrefixRegex(t *testing.T) {
	prefix, comment, success := GetCommentAfterPrefixRegex("//ProductController 产品", "ProductController")
	assert.Equal(t, prefix, "ProductController")
	assert.Equal(t, success, true)
	assert.Equal(t, comment, "产品")
	prefix, comment, success = GetCommentAfterPrefixRegex("//ProductController 产品,anytexthere", "ProductController")
	assert.Equal(t, prefix, "ProductController")
	assert.Equal(t, success, true)
	assert.Equal(t, comment, "产品,anytexthere")
	prefix, comment, success = GetCommentAfterPrefixRegex("//ProductController 产品    anytexthere", "ProductController")
	assert.Equal(t, prefix, "ProductController")
	assert.Equal(t, success, true)
	assert.Equal(t, comment, "产品    anytexthere")
	prefix, comment, success = GetCommentAfterPrefixRegex("//ProductController    产品    anytexthere", "ProductController")
	assert.Equal(t, prefix, "ProductController")
	assert.Equal(t, success, true)
	assert.Equal(t, comment, "产品    anytexthere")
	prefix, comment, success = GetCommentAfterPrefixRegex("//@TAG /ddd    产品    anytexthere", "")
	assert.Equal(t, prefix, "@TAG")
	assert.Equal(t, success, true)
	assert.Equal(t, comment, "/ddd    产品    anytexthere")

}
