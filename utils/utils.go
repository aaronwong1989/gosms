package utils

type OctetString struct {
	S string
}

func (s *OctetString) ToBytes(length int) []byte {
	if len(s.S) == length {
		return []byte(s.S)
	} else if len(s.S) < length {
		nb := []byte(s.S)
		for i := len(s.S); i < length; i++ {
			nb = append(nb, 0)
		}
		return nb
	} else {
		return []byte(s.S)[:length]
	}
}
