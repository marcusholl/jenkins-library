package transportrequest

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type commitIteratorMock struct {
	commits []object.Commit
	index   int
}

func (iter *commitIteratorMock) Next() (*object.Commit, error) {
	i := iter.index
	iter.index++

	if i >= len(iter.commits) {
		return nil, io.EOF // real iterators also behave like this
	}
	return &iter.commits[i], nil
}

func (iter *commitIteratorMock) ForEach(cb func(c *object.Commit) error) error {
	for {
		c, err := iter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = cb(c)
		if err == storer.ErrStop {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (iter *commitIteratorMock) Close() {

}

func TestRetrieveLabelStraightForward(t *testing.T) {

	t.Run("single commit tests", func(t * testing.T) {

		runTest := func(message string, asserts func([]string)) {
			commitIter := &commitIteratorMock{
				commits: []object.Commit{
					object.Commit{
						Hash:    plumbing.NewHash("3434343434343434343434343434343434343434"),
						Message: message,
					},
				},
			}
			labels, err := FindLabelsInCommits(commitIter, "TransportRequest")
			if assert.NoError(t, err) {
				asserts(labels)
			}
		}
		_ = runTest
		t.Run("straight forward", func(t * testing.T) {
			runTest(
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678",
				func(labels []string) {
					assert.Equal(t, []string{"12345678"}, labels)
				},
			)
		})
		t.Run("trailing spaces after our value", func(t * testing.T) {
			runTest(
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678  ",
				func(labels []string) {
					assert.Equal(t, []string{"12345678"}, labels)
				},
			)
		})
		t.Run("trailing text after our value", func(t * testing.T) {
			runTest(
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\nTransportRequest: 12345678 aaa",
				func(labels []string) {
					assert.Empty(t, labels)
				},
			)
		})
		t.Run("Leading whitespace before our label", func(t * testing.T) {
			runTest(
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\n   TransportRequest: 12345678",
				func(labels []string) {
					assert.Equal(t, []string{"12345678"}, labels)
				},
			)
		})
		t.Run("leading text before our label", func(t * testing.T) {
			runTest(
				"this is a commit with TransportRequestId\n\nThis is the first line of the message body\naaa TransportRequest: 12345678",
				func(labels []string) {
					assert.Empty(t, labels)
				},
			)
		})


	})
}
