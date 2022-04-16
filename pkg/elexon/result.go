package elexon

type ElexonAPIResult struct {
	data map[string]interface{}
}

func (ar ElexonAPIResult) Int(name string) int {
	v, ck := ar.data[name]
	if !ck {
		return -1
	}
	return v.(int)
}

func (ar ElexonAPIResult) String(name string) string {
	v, ck := ar.data[name]
	if !ck {
		return ""
	}
	return v.(string)
}

func (ar ElexonAPIResult) Bool(name string) bool {
	v, ck := ar.data[name]
	if !ck {
		return false
	}
	return v.(bool)
}
