package commands

import (
	"errors"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
)

// Fetch is a FETCH command, as defined in RFC 3501 section 6.4.5.
type Fetch struct {
	SeqSet       *imap.SeqSet     //
	Items        []imap.FetchItem //
	modSeqFlag   bool             // если данный флаг установлен в true надо анализировать ChangedSince
	ChangedSince uint32           // число которое передается в ChangedSince
}

func (cmd *Fetch) Command() *imap.Command {
	items := make([]interface{}, len(cmd.Items))
	for i, item := range cmd.Items {
		items[i] = imap.RawString(item)
	}

	return &imap.Command{
		Name:      "FETCH",
		Arguments: []interface{}{cmd.SeqSet, items},
	}
}

// getChangedSince разбираем параметры вида [CHANGEDSINCE 0]
func (cmd *Fetch) getChangedSince(param string) {
	if cmd.modSeqFlag {
		changeSince, _ := strconv.ParseUint(param, 10, 32)
		cmd.ChangedSince = uint32(changeSince)
	} else {
		if strings.ToUpper(param) == "CHANGEDSINCE" {
			cmd.Items = append(cmd.Items, imap.FetchModSeq)
		}
	}
}

func (cmd *Fetch) Parse(fields []interface{}) error {
	if len(fields) < 2 {
		return errors.New("No enough arguments")
	}

	var err error
	if seqset, ok := fields[0].(string); !ok {
		return errors.New("Sequence set must be an atom")
	} else if cmd.SeqSet, err = imap.ParseSeqSet(seqset); err != nil {
		return err
	}

	switch items := fields[1].(type) {
	case string: // A macro or a single item
		cmd.Items = imap.FetchItem(strings.ToUpper(items)).Expand()
	case []interface{}: // A list of items
		cmd.Items = make([]imap.FetchItem, 0, len(items))
		for _, v := range items {
			itemStr, _ := v.(string)
			item := imap.FetchItem(strings.ToUpper(itemStr))
			cmd.Items = append(cmd.Items, item.Expand()...)
		}
	default:
		return errors.New("Items must be either a string or a list")
	}

	if len(fields) > 2 {
		switch items := fields[2].(type) {
		case string:
			cmd.getChangedSince(items)
		case []interface{}:
			for _, i := range items {
				cmd.getChangedSince(i.(string))
			}
		default:
			return errors.New("Items must be either a string or a list")
		}
	}
	return nil
}
