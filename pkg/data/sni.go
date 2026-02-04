package data

import (
	"encoding/binary"
)

// ExtractSNI extracts the Server Name Indication from a TLS ClientHello
func ExtractSNI(payload []byte) string {
	if len(payload) < 5 {
		return ""
	}

	// Check for TLS handshake record (0x16)
	if payload[0] != 0x16 {
		return ""
	}

	// TLS version (we accept 1.0 through 1.3)
	// 0x0301 = TLS 1.0, 0x0302 = TLS 1.1, 0x0303 = TLS 1.2/1.3
	version := binary.BigEndian.Uint16(payload[1:3])
	if version < 0x0301 || version > 0x0303 {
		return ""
	}

	// Record length
	recordLen := int(binary.BigEndian.Uint16(payload[3:5]))
	if recordLen+5 > len(payload) {
		return ""
	}

	// Handshake message starts at offset 5
	handshake := payload[5:]
	if len(handshake) < 4 {
		return ""
	}

	// Check for ClientHello (type 0x01)
	if handshake[0] != 0x01 {
		return ""
	}

	// Handshake length (3 bytes, big endian)
	handshakeLen := int(handshake[1])<<16 | int(handshake[2])<<8 | int(handshake[3])
	if handshakeLen+4 > len(handshake) {
		return ""
	}

	// ClientHello body starts at offset 4
	clientHello := handshake[4:]
	if len(clientHello) < 38 {
		return ""
	}

	// Skip version (2 bytes) + random (32 bytes)
	offset := 34

	// Session ID length (1 byte) + session ID
	if offset >= len(clientHello) {
		return ""
	}
	sessionIDLen := int(clientHello[offset])
	offset++
	offset += sessionIDLen
	if offset >= len(clientHello) {
		return ""
	}

	// Cipher suites length (2 bytes) + cipher suites
	if offset+2 > len(clientHello) {
		return ""
	}
	cipherSuitesLen := int(binary.BigEndian.Uint16(clientHello[offset:]))
	offset += 2 + cipherSuitesLen
	if offset >= len(clientHello) {
		return ""
	}

	// Compression methods length (1 byte) + compression methods
	if offset >= len(clientHello) {
		return ""
	}
	compMethodsLen := int(clientHello[offset])
	offset++
	offset += compMethodsLen
	if offset >= len(clientHello) {
		return ""
	}

	// Extensions length (2 bytes)
	if offset+2 > len(clientHello) {
		return ""
	}
	extensionsLen := int(binary.BigEndian.Uint16(clientHello[offset:]))
	offset += 2
	extensionsEnd := offset + extensionsLen

	// Parse extensions looking for SNI (type 0x0000)
	for offset+4 <= extensionsEnd && offset+4 <= len(clientHello) {
		extType := binary.BigEndian.Uint16(clientHello[offset:])
		extLen := int(binary.BigEndian.Uint16(clientHello[offset+2:]))
		offset += 4

		if offset+extLen > len(clientHello) {
			return ""
		}

		// SNI extension type is 0
		if extType == 0x0000 {
			return parseSNIExtension(clientHello[offset : offset+extLen])
		}

		offset += extLen
	}

	return ""
}

func parseSNIExtension(data []byte) string {
	if len(data) < 5 {
		return ""
	}

	// SNI list length (2 bytes)
	listLen := int(binary.BigEndian.Uint16(data))
	if listLen+2 > len(data) {
		return ""
	}

	offset := 2
	for offset+3 <= len(data) && offset < listLen+2 {
		nameType := data[offset]
		nameLen := int(binary.BigEndian.Uint16(data[offset+1:]))
		offset += 3

		if offset+nameLen > len(data) {
			return ""
		}

		// Type 0 is hostname
		if nameType == 0 {
			hostname := string(data[offset : offset+nameLen])
			// Validate hostname (basic check)
			if len(hostname) > 0 && len(hostname) < 256 {
				return hostname
			}
		}

		offset += nameLen
	}

	return ""
}

// IsTLSClientHello checks if payload looks like a TLS ClientHello
func IsTLSClientHello(payload []byte) bool {
	if len(payload) < 6 {
		return false
	}
	// Record type 0x16 (handshake), version 0x0301-0x0303, handshake type 0x01 (ClientHello)
	return payload[0] == 0x16 &&
		payload[1] == 0x03 &&
		payload[2] >= 0x01 && payload[2] <= 0x03 &&
		payload[5] == 0x01
}
