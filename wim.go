package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

// WIM was generated 2023-02-04 19:27:11 by https://xml-to-go.github.io/ in Ukraine.
type WIM struct {
	Path       string
	XMLName    xml.Name `xml:"WIM"`
	Text       string   `xml:",chardata"`
	TOTALBYTES struct {
		Text string `xml:",chardata"`
	} `xml:"TOTALBYTES"`
	IMAGE []struct {
		Summary  string
		Text     string `xml:",chardata"`
		INDEX    string `xml:"INDEX,attr"`
		DIRCOUNT struct {
			Text string `xml:",chardata"`
		} `xml:"DIRCOUNT"`
		FILECOUNT struct {
			Text string `xml:",chardata"`
		} `xml:"FILECOUNT"`
		TOTALBYTES struct {
			Text string `xml:",chardata"`
		} `xml:"TOTALBYTES"`
		HARDLINKBYTES struct {
			Text string `xml:",chardata"`
		} `xml:"HARDLINKBYTES"`
		CREATIONTIME struct {
			Text     string `xml:",chardata"`
			HIGHPART struct {
				Text string `xml:",chardata"`
			} `xml:"HIGHPART"`
			LOWPART struct {
				Text string `xml:",chardata"`
			} `xml:"LOWPART"`
		} `xml:"CREATIONTIME"`
		LASTMODIFICATIONTIME struct {
			Text     string `xml:",chardata"`
			HIGHPART struct {
				Text string `xml:",chardata"`
			} `xml:"HIGHPART"`
			LOWPART struct {
				Text string `xml:",chardata"`
			} `xml:"LOWPART"`
		} `xml:"LASTMODIFICATIONTIME"`
		WIMBOOT struct {
			Text string `xml:",chardata"`
		} `xml:"WIMBOOT"`
		WINDOWS struct {
			Text string `xml:",chardata"`
			ARCH struct {
				Text string `xml:",chardata"`
			} `xml:"ARCH"`
			PRODUCTNAME struct {
				Text string `xml:",chardata"`
			} `xml:"PRODUCTNAME"`
			EDITIONID struct {
				Text string `xml:",chardata"`
			} `xml:"EDITIONID"`
			INSTALLATIONTYPE struct {
				Text string `xml:",chardata"`
			} `xml:"INSTALLATIONTYPE"`
			SERVICINGDATA struct {
				Text          string `xml:",chardata"`
				GDRDUREVISION struct {
					Text string `xml:",chardata"`
				} `xml:"GDRDUREVISION"`
				PKEYCONFIGVERSION struct {
					Text string `xml:",chardata"`
				} `xml:"PKEYCONFIGVERSION"`
			} `xml:"SERVICINGDATA"`
			HAL struct {
				Text string `xml:",chardata"`
			} `xml:"HAL"`
			PRODUCTTYPE struct {
				Text string `xml:",chardata"`
			} `xml:"PRODUCTTYPE"`
			PRODUCTSUITE struct {
				Text string `xml:",chardata"`
			} `xml:"PRODUCTSUITE"`
			LANGUAGES struct {
				Text     string `xml:",chardata"`
				LANGUAGE struct {
					Text string `xml:",chardata"`
				} `xml:"LANGUAGE"`
				DEFAULT struct {
					Text string `xml:",chardata"`
				} `xml:"DEFAULT"`
			} `xml:"LANGUAGES"`
			VERSION struct {
				Text  string `xml:",chardata"`
				MAJOR struct {
					Text string `xml:",chardata"`
				} `xml:"MAJOR"`
				MINOR struct {
					Text string `xml:",chardata"`
				} `xml:"MINOR"`
				BUILD struct {
					Text string `xml:",chardata"`
				} `xml:"BUILD"`
				SPBUILD struct {
					Text string `xml:",chardata"`
				} `xml:"SPBUILD"`
				SPLEVEL struct {
					Text string `xml:",chardata"`
				} `xml:"SPLEVEL"`
			} `xml:"VERSION"`
			SYSTEMROOT struct {
				Text string `xml:",chardata"`
			} `xml:"SYSTEMROOT"`
		} `xml:"WINDOWS"`
		NAME struct {
			Text string `xml:",chardata"`
		} `xml:"NAME"`
		DESCRIPTION struct {
			Text string `xml:",chardata"`
		} `xml:"DESCRIPTION"`
		FLAGS struct {
			Text string `xml:",chardata"`
		} `xml:"FLAGS"`
		DISPLAYNAME struct {
			Text string `xml:",chardata"`
		} `xml:"DISPLAYNAME"`
		DISPLAYDESCRIPTION struct {
			Text string `xml:",chardata"`
		} `xml:"DISPLAYDESCRIPTION"`
	} `xml:"IMAGE"`
}

func DecodeUTF16(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}

func WIMInfo(wim string) (*WIM, error) {
	_, err := Run("wiminfo", "--extract-xml="+wim+"-info.xml", wim)
	if err != nil {
		return nil, err
	}
	defer os.Remove(wim + "-info.xml")

	wimxmlb, err := ioutil.ReadFile(wim + "-info.xml")
	if err != nil {
		return nil, err
	}

	wimxml, err := DecodeUTF16(wimxmlb)
	if err != nil {
		return nil, err
	}

	wiminfo := &WIM{Path: wim}
	err = xml.Unmarshal([]byte(wimxml), wiminfo)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(wiminfo.IMAGE); i++ {
		img := wiminfo.IMAGE[i]
		edition := img.NAME.Text
		if img.DESCRIPTION.Text != "" && img.DESCRIPTION.Text != img.NAME.Text {
			edition += " - " + img.DESCRIPTION.Text
		}
		edition += " (NT" + img.WINDOWS.VERSION.MAJOR.Text + "." + img.WINDOWS.VERSION.MINOR.Text + " Build " + img.WINDOWS.VERSION.BUILD.Text
		if img.WINDOWS.VERSION.SPLEVEL.Text != "" && img.WINDOWS.VERSION.SPLEVEL.Text != "0" {
			edition += " Service Pack " + img.WINDOWS.VERSION.SPLEVEL.Text + " v" + img.WINDOWS.VERSION.SPBUILD.Text
		}
		edition += ")"
		wiminfo.IMAGE[i].Summary = edition
	}

	return wiminfo, nil
}
