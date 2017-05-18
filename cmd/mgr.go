// Copyright © 2017 edwin <edwin.lzh@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"go.uber.org/zap"
	mgo "gopkg.in/mgo.v2"

	"github.com/lvzhihao/wechat/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mgrCmd represents the mgr command
var mgrCmd = &cobra.Command{
	Use:   "mgr",
	Short: "manager wechat config",
	Long:  `manager wechat config`,
	Run: func(cmd *cobra.Command, args []string) {
		// mongo dial
		session, err := mgo.Dial(viper.GetString("mongo"))
		if err != nil {
			log.Panic("mongo config error", zap.Error(err))
		}
		defer session.Close()
		// ensure index
		models.WechatConfigEnusreConfig(session)

		if len(args) == 1 {
			switch args[0] {
			case "import":
				importConfig(session)
			case "export":
				exportConfig(session)
			default:
				listConfig(session)
			}
		} else {
			listConfig(session)
		}
	},
}

func exportConfig(s *mgo.Session) {
	return
}

func importConfig(s *mgo.Session) {
	b, err := ioutil.ReadFile("wechat.json")
	if err != nil {
		log.Fatal(err)
	} else {
		var configs []models.WechatConfig
		err := json.Unmarshal(b, &configs)
		if err != nil {
			log.Fatal(err)
		} else {
			for _, config := range configs {
				config.Create(s)
				config.DetailEnsure(s)
			}
		}
	}
}

func listConfig(s *mgo.Session) {
	list, err := models.ListWechatConfig(s)
	if err != nil {
		log.Printf("ShowAll Error: %s\n", err)
	} else {
		for _, config := range list {
			log.Println(config)
		}
	}
}

func init() {
	RootCmd.AddCommand(mgrCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// mgrCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// mgrCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	mgrCmd.Flags().String("mongo", "mongodb://127.0.0.1/wechat", "mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]")
	viper.BindPFlag("mongo", mgrCmd.Flags().Lookup("mongo"))
}
