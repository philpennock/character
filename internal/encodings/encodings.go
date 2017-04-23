// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package encodings

import (
	"fmt"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

type unknownCharsetError struct {
	input string
}

func (u unknownCharsetError) Error() string {
	return fmt.Sprintf("unknown charset: %q", u.input)
}

// ListKnownCharsets returns a slice of strings representing the names which we
// advertise as being known.  We support a few more, as aliases, but they're not
// reported.
func ListKnownCharsets() []string {
	return []string{
		"UTF-8",
		// sort the rest
		"Big5",
		"EUC-JP",
		"EUC-KR",
		"GB18030",
		"GBK",
		"HZ-GB2312",
		"ISO-2022-JP",
		"Shift-JIS",
	}
}

// LoadCharsetDecoder returns an encoding decoder matching the given name; it
// is designed to take user input, so tries to normalize and handle various
// aliases.
func LoadCharsetDecoder(charset string) (*encoding.Decoder, error) {
	if charset == "" {
		return encoding.Nop.NewDecoder(), nil
	}
	lcShort := strings.Replace(strings.ToLower(charset), "-", "", -1)
	switch lcShort {
	case "utf8", "unicode":
		return encoding.Nop.NewDecoder(), nil
	case "eucjp":
		return japanese.EUCJP.NewDecoder(), nil
	case "euckr", "cp949":
		return korean.EUCKR.NewDecoder(), nil
	case "iso2022jp":
		return japanese.ISO2022JP.NewDecoder(), nil
	case "shiftjis", "cp932", "windows31j":
		return japanese.ShiftJIS.NewDecoder(), nil
	case "gb18030":
		return simplifiedchinese.GB18030.NewDecoder(), nil
	case "gbk", "cp936":
		return simplifiedchinese.GBK.NewDecoder(), nil
	case "hzgb2312":
		return simplifiedchinese.HZGB2312.NewDecoder(), nil
	case "big5", "cp950":
		return traditionalchinese.Big5.NewDecoder(), nil
	}

	return nil, unknownCharsetError{input: charset}
}
