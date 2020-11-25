package responses

import (
	"fmt"

	"github.com/emersion/go-imap"
)

const searchName = "SEARCH"

// A SEARCH response.
// See RFC 3501 section 7.2.5
type Search struct {
	ReturnValue string
	Tag         string
	Ids         []uint32
}

func (r *Search) Handle(resp imap.Resp) error {
	name, fields, ok := imap.ParseNamedResp(resp)
	if !ok || name != searchName {
		return ErrUnhandled
	}

	r.Ids = make([]uint32, len(fields))
	for i, f := range fields {
		if id, err := imap.ParseNumber(f); err != nil {
			return err
		} else {
			r.Ids[i] = id
		}
	}

	return nil
}

func (r *Search) WriteTo(w *imap.Writer) (err error) {
	var fields []interface{}
	if len(r.ReturnValue) > 0 {
		res := fmt.Sprintf("ESEARCH (TAG \"%s\") UID ", r.Tag)
		if r.ReturnValue == "COUNT" {
			if len(r.Ids) > 0 {
				res += fmt.Sprintf("COUNT %d", r.Ids[0])
			}
		} else if r.ReturnValue == "ALL" {
			seq := imap.SeqSet{}
			seq.AddNum(r.Ids...)
			res += fmt.Sprintf("ALL %s", seq.String())
		}
		fields = []interface{}{imap.RawString(res)}

	} else {
		fields = []interface{}{imap.RawString(searchName)}
		for _, id := range r.Ids {
			fields = append(fields, id)
		}
	}
	resp := imap.NewUntaggedResp(fields)
	return resp.WriteTo(w)
}
