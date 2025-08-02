package key

import "crypto/rand"

func GenerateIdentifier(len uint16) []byte {
	token := make([]byte, len)
    rand.Read(token)
    
	return token
} 