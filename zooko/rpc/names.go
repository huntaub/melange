package rpc

import "github.com/melange-app/nmcd/btcjson"

func (r *Server) LookupName(name string) (string, bool, error) {
	cmd, err := btcjson.NewNameShowCmd(nil, name)
	if err != nil {
		return "", false, err
	}

	reply, err := r.Send(cmd)
	if err != nil {
		return "", false, err
	}

	if reply.Result == nil {
		return "", false, errNilReply
	} else if reply.Error != nil {
		// A code of -4 indicates that the name was not found
		// in the database.
		if reply.Error.Code == -4 {
			return "", false, nil
		}

		return "", false, reply.Error
	}

	nameInfo := reply.Result.(*btcjson.NameInfoResult)

	if nameInfo.Expired {
		return "", false, nil
	}

	// Return the name
	return nameInfo.Value, true, nil
}
