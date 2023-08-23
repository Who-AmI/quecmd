package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"quecmd/utils"
	"strings"
	"sync"
	"syscall"

	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cobra"
)

var (
	fileQue     string
	conNum      int
	cmdStr      string
	target      string
	shareScreen string
	queryAll    string
	dbpath      string
)

var rootCmd = &cobra.Command{
	Use:   "quecmd",
	Short: "create queue and execute cmd concurrent",
}

var subQueue = &cobra.Command{
	Use:   "queue",
	Short: "create queue and run command",
	Run:   RunQue,
}
var subAdd = &cobra.Command{
	Use:   "add",
	Short: "add task to queue",
	Run:   Add2Que,
}
var subQuery = &cobra.Command{
	Use:   "query",
	Short: "query tasks status",
	Run:   QueryTaskStatus,
}
var subInit = &cobra.Command{
	Use:   "init",
	Short: "init db",
	Run:   InitDB,
}
var subVersion = &cobra.Command{
	Use:   "version",
	Short: "show version",
	Run:   PrintVerion,
}

func InitDB(_ *cobra.Command, args []string) {

	utils.InitSqliteDB(dbpath)
	return
}

func RunQue(_ *cobra.Command, args []string) {

	fmt.Fprintf(os.Stderr, "RunQue...")
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(conNum, func(i interface{}) {
		DealTask(i.(string))
		wg.Done()

	}, ants.WithPreAlloc(true))
	defer p.Release()
	os.Remove(fileQue)
	err := syscall.Mkfifo(fileQue, 0666)
	if err != nil {
		panic(err)
	}
	pipe, err := os.OpenFile(fileQue, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		panic(err)
	}
	defer pipe.Close()
	reader := bufio.NewReader(pipe)
	defer os.Remove(fileQue)
	go func() {

		for {
			line, err := reader.ReadBytes('\n')
			if err == nil {
				l := strings.Replace(string(line), "\n", "", -1)
				fmt.Fprintf(os.Stderr, "New target: [%s]\n", l)

				wg.Add(1)
				p.Invoke(l)
			}

		}

	}()

	<-make(chan struct{})

}
func DealTask(t string) {

	var args string = ""

	cmdStr = strings.TrimSpace(cmdStr)
	arr := strings.Split(cmdStr, " ")
	arr = append(arr, t)
	if len(arr) > 2 {
		args = strings.Join(arr[1:len(arr)-1], " ")
	}
	cmd := exec.Command(arr[0], arr[1:]...)

	fmt.Fprintf(os.Stderr, "cmd:[%s],target:[%s],args:[%s]\n", arr[0], t, args)
	if shareScreen == "true" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	utils.UpdateTaskStatus(t, "finished", dbpath)
}
func Add2Que(_ *cobra.Command, args []string) {

	is_submit, _ := utils.CheckTaskIsRunning(target, dbpath)
	if is_submit {
		fmt.Fprintf(os.Stderr, "target %s is runing; skip!\n", target)
		return
	}

	f, err := os.OpenFile(fileQue, os.O_RDWR, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	n, err := f.WriteString(target + "\n")
	if err != nil {
		fmt.Println(err)
		return
	}
	if n != (len(target) + 1) {
		fmt.Printf("error: write %d byte ", n)
		return
	}
	err = utils.InsertTask(target, dbpath)
	if err != nil {
		fmt.Println(err)
		return

	}

}
func PrintVerion(_ *cobra.Command, args []string) {

	fmt.Println("author:wuwuwu, version: 1.0")

}
func QueryTaskStatus(_ *cobra.Command, args []string) {

	utils.QueryAllTasks(target, dbpath)
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.AddCommand(subQueue, subAdd, subQuery, subVersion, subInit)

	subInit.Flags().StringVarP(&dbpath, "dbpath", "d", "./.quecmd.db", "sqlite3 db file")

	subQueue.Flags().StringVarP(&fileQue, "queue", "q", "./queue.que", "")
	subQueue.Flags().StringVarP(&cmdStr, "run", "r", "echo hello", "")
	subQueue.Flags().IntVarP(&conNum, "concurrency", "c", 4, "")
	subQueue.Flags().StringVarP(&shareScreen, "share", "s", "true", "share screen output")

	subQuery.Flags().StringVarP(&target, "target", "t", "", "query a task status")
	subQuery.Flags().StringVarP(&dbpath, "dbpath", "d", "./.quecmd.db", "sqlite3 db file")

	subAdd.Flags().StringVarP(&target, "target", "t", "", "add target")
	subAdd.Flags().StringVarP(&fileQue, "queue", "q", "./queue.que", "")

}
