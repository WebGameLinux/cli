package resolver

import (
		"os"
		"strings"
)

type IniParser struct {
		FileName string
		IM       map[string]map[string]string
}

const ROOT = "global"

func (iniP *IniParser) ParserIniFile(fileName string) error {
		var (
				err      error
				line     string
				file     *os.File
				partName = ROOT
				buf      = make([]byte, 1)
		)

		iniP.FileName = fileName
		iniP.IM = map[string]map[string]string{}
		if file, err = os.Open(iniP.FileName); err != nil {
				return err
		}
		iniP.IM[partName] = map[string]string{}
		for n, err := file.Read(buf); n > 0; {
				if err != nil {
						return err
				}
				if string(buf) != "\n" {
						line += string(buf)
				} else {
						iCode, str := parserIniLine(line)
						switch iCode {
						case 1:
								partName = str
								iniP.IM[partName] = map[string]string{}
						case 2:
								aName := strings.Split(str, "=")[0]
								aValue := strings.Split(str, "=")[1]

								iniP.IM[partName][aName] = aValue

						}
						line = ""
				}
				if n, err = file.Read(buf); err != nil {
						if err.Error() != "EOF" {
								return err
						}
						break
				}
		}
		return nil
}

func parserIniLine(line string) (int, string) {
		cmtIdx := strings.Index(line, ";")
		if cmtIdx > 0 {
				line = strings.Split(line, ";")[0]
		}
		bracketsLeft := strings.Index(line, "[")
		bracketsRight := strings.Index(line, "]")

		if bracketsLeft >= 0 && bracketsRight > bracketsLeft {
				line = strings.Split(line, "[")[1]
				line = strings.Split(line, "]")[0]
				return 1, line
		}
		if bracketsLeft < 0 && bracketsRight < 0 && strings.Contains(line, "=") {
				return 2, line
		}
		return 0, ""
}

func (iniP *IniParser) GetName(pName string, vName string) string {
		return iniP.IM[pName][vName]
}

func NewIniParser() *IniParser {
		return new(IniParser)
}
