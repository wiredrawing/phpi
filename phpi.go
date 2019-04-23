// +build windows

package main

import (
	"bufio"
	_ "errors"
	"fmt"
	"io"
	_ "io"
	"io/ioutil"
	"os"
	exe "os/exec"
	"os/signal"
	"path/filepath"
	. "phpi/echo"
	"phpi/goroutine"
	"phpi/standardInput"
	"reflect"
	_ "reflect"
	_ "regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	_ "strings"
	_ "syscall"
	_ "time"

	// 自作パッケージ

	_ "phpi/myreflect"

	"golang.org/x/sys/windows"

	// syscallライブラリの代替ツール

	_ "golang.org/x/sys/unix"
)

func enClosure(keep string) func(string) string {

	var localVariable string = keep
	return func(param string) string {
		if len(param) > 0 {
			localVariable = param
			return localVariable
		} else {
			return localVariable
		}
	}

}

// 空のインターフェース
type TestInterface interface {
	Fuck()
}

type TestStruct struct {
	Name    string
	Fucking string
}

func (this *TestStruct) Fuck() {
	fmt.Println("Fuck you")
}
func main() {
	var oo TestStruct = TestStruct{}
	var io TestInterface = &oo
	// この時点でioは*TestStruct型になる
	fmt.Println(reflect.TypeOf(io))
	func(i interface{}) {
		value, ok := io.(*TestStruct)
		if ok == true {
			value.Fuck()
		}
	}(io)

	var echo func(interface{}) (int, error)
	echo = Echo()
	var stdin (func(*string) bool) = nil
	var standard *standardInput.StandardInput = new(standardInput.StandardInput)
	standard.SetStandardInputFunction()
	standard.SetBufferSize(1024 * 2)
	stdin = standard.GetStandardInputFunction()

	// プロセスの監視
	var signal_chan chan os.Signal = make(chan os.Signal)
	// OSによってシグナルのパッケージを変更
	signal.Notify(
		signal_chan,
		os.Interrupt,
		os.Kill,
		windows.SIGKILL,
		windows.SIGHUP,
		windows.SIGINT,
		windows.SIGTERM,
		windows.SIGQUIT,
		windows.Signal(0x13),
		windows.Signal(0x14), // Windowsの場合 SIGTSTPを認識しないためリテラルで指定する
	)

	// シグナルを取得後終了フラグとするチャンネル
	var exit_chan chan int = make(chan int)
	// シグナルを監視
	go goroutine.MonitoringSignal(signal_chan, exit_chan)
	// コンソールを停止するシグナルを握りつぶす
	go goroutine.CrushingSignal(exit_chan)
	// 平行でGCを実施
	go goroutine.RunningFreeOSMemory()

	// 実行するPHPスクリプトの初期化
	// バックティックでヒアドキュメント
	const initializer = "<?php \r\n" +
		"ini_set(\"display_errors\", 1);\r\n" +
		"ini_set(\"error_reporting\", -1);\r\n"

	// 利用変数初期化
	var input string
	var line *string
	line = new(string)

	var tentativeFile *string
	tentativeFile = new(string)

	var writtenByte *int
	writtenByte = new(int)

	var ff *os.File
	var err error
	// ダミー実行ポインタ
	ff, err = ioutil.TempFile("", "__php__main__")
	if err != nil {
		echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	ff.Chmod(os.ModePerm)
	*writtenByte, err = ff.WriteAt([]byte(initializer), 0)
	if err != nil {
		echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	// ファイルポインタに書き込まれたバイト数を検証する
	if *writtenByte != len(initializer) {
		echo("[Couldn't complete process to initialize script file.]\r\n")
		os.Exit(255)
	}
	// ファイルポインタオブジェクトから絶対パスを取得する
	*tentativeFile, err = filepath.Abs(ff.Name())
	if err != nil {
		echo(err.Error() + "\r\n")
		os.Exit(255)
	}
	defer ff.Close()
	defer os.Remove(*tentativeFile)

	var count int = 0
	//var ss int = 0
	var multiple int = 0
	//var backup []byte = make([]byte, 0)
	var currentDir string

	// saveコマンド入力用
	var saveFp *os.File
	saveFp = new(os.File)

	// 入力されたソースコードをバックグラウンドで検証する
	var syntax chan int
	syntax = make(chan int)
	var cc chan int
	cc = make(chan int)

	var fixedInput string
	input = initializer
	fixedInput = input
	var exitCode int

	// channelの代替方法
	// var wg *sync.WaitGroup = new(sync.WaitGroup)
	for {
		if multiple == 1 {
			echo("(" + strconv.Itoa(exitCode) + ")" + " .... ")
		} else {
			echo("(" + strconv.Itoa(exitCode) + ")" + "php > ")
		}
		*line = ""

		// 標準入力開始
		stdin(line)
		temp := *line

		if temp == "del" {
			ff, err = deleteFile(ff, initializer)
			if err != nil {
				echo(err.Error() + "\r\n")
				os.Exit(255)
			}
			*line = ""
			input = initializer
			fixedInput = input
			count = 0
			multiple = 0
			continue
		} else if temp == "save" {
			currentDir, err = os.Getwd()
			currentDir += "\\save.php"
			saveFp, err = os.Create(currentDir)
			if err != nil {
				echo(err.Error() + "\r\n")
				continue
			}
			saveFp.Chmod(os.ModePerm)
			input = fixedInput
			*writtenByte, err = saveFp.WriteAt([]byte(input), 0)
			if err != nil {
				saveFp.Close()
				echo(err.Error() + "\r\n")
				os.Exit(255)
			}
			echo("[" + currentDir + ":Completed saving input code which you wrote.]" + "\r\n")
			saveFp.Close()
			*line = ""
			multiple = 0
			exitCode = 0
			continue
		} else if temp == "exit" {
			// コンソールを終了させる
			echo("[Would you really like to quit a console which you are running in terminal? yes or other]\r\n")
			var quitText *string
			quitText = new(string)
			stdin(quitText)
			if *quitText == "yes" {
				os.Exit(0)
			} else {
				echo("[Canceled to quit this console app in terminal.]\r\n")
			}
			*line = ""
			continue
		} else if temp == "restore" || temp == "clear" {
			input = fixedInput
			os.Truncate(*tentativeFile, 0)
			ff.WriteAt([]byte(input), 0)
			multiple = 0
			exitCode = 0
			continue
		} else if temp == "" {
			// 空文字エンターの場合はループを飛ばす
			continue
		}

		input += *line + "\n"

		_, err = ff.WriteAt([]byte(input), 0)
		if err != nil {
			// temporary fileへの書き込みに失敗した場合
			echo(err.Error())
			continue
		}
		// 並行処理でスクリプトが正常実行できるまでループを繰り返す
		// wg.Add(1)
		go SyntaxCheck(tentativeFile, syntax, cc /*, wg*/)
		// チャンネルから値を取得
		si := <-syntax
		exitCode = <-cc
		//		wg.Wait()
		if si == 1 {
			*line = ""
			fixedInput = input + "echo (PHP_EOL);"
			count, err = tempFunction(ff, tentativeFile, count, false)
			if err != nil {
				echo(err.Error())
				continue
			}
			multiple = 0
			input += " echo(PHP_EOL);\r\n "
		} else {
			_, err = tempFunction(ff, tentativeFile, count, true)
			multiple = 1
		}
	}
}

func SyntaxCheck(filePath *string, c chan int, cc chan int /*wg *sync.WaitGroup*/) (bool, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	var e error = nil
	var command *exe.Cmd
	// バックグラウンドでPHPをコマンドラインで実行
	command = exe.Command("php", *filePath)
	e = command.Run()
	//wg.Done()
	if e == nil {
		// コマンド成功時
		c <- 1
		cc <- command.ProcessState.ExitCode()
		return true, nil
	} else {
		// コマンド実行失敗時
		c <- 0
		cc <- command.ProcessState.ExitCode()
		return false, e
	}
}

func tempFunction(fp *os.File, filePath *string, beforeOffset int, errorCheck bool) (int, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	echo := Echo()
	var e error
	var stdout io.ReadCloser
	var command *exe.Cmd
	command = exe.Command("php", *filePath)
	// バックグラウンドでPHPをコマンドラインで実行
	e = command.Run()

	if errorCheck == true {
		// バックグランドでの実行が失敗の場合
		if e != nil {
			// 実行したスクリプトの終了コードを取得
			var code bool = command.ProcessState.Success()
			if code != true {
				var scanText string = ""
				command = exe.Command("php", *filePath)
				stdout, _ := command.StdoutPipe()
				command.Start()
				scanner := bufio.NewScanner(stdout)
				var ii int = 0
				for scanner.Scan() {
					if ii >= beforeOffset {
						scanText = scanner.Text()
						if len(scanText) > 0 {
							echo("     " + scanner.Text() + "\r\n")
						}
					}
					ii++
				}
				if beforeOffset > ii {
					command = exe.Command("php", *filePath)
					stdout, _ := command.StdoutPipe()
					command.Start()
					scanner = bufio.NewScanner(stdout)
					for scanner.Scan() {
						scanText = scanner.Text()
						if len(scanText) > 0 {
							echo("     " + scanner.Text() + "\r\n")
						}
					}
				}
				command.Wait()
				echo("\r\n")
				command = nil
				stdout = nil
				return beforeOffset, e
			}
		}
	}
	var ii int = 0
	var scanText string
	// Run()メソッドで利用したcommandオブジェクトを再利用
	command = exe.Command("php", *filePath)
	stdout, e = command.StdoutPipe()
	if e != nil {
		echo(e.Error() + "\r\n")
		panic("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus.")
	}
	command.Start()
	scanner := bufio.NewScanner(stdout)
	for {
		// 読み取り可能な場合
		if scanner.Scan() == true {
			if ii >= beforeOffset {
				scanText = scanner.Text()
				if len(scanText) > 0 {
					echo("     " + scanText + "\r\n")
				}
			}
			ii++
		} else {
			break
		}
	}
	command.Wait()
	command = nil
	stdout = nil
	scanText = ""
	echo("\r\n")
	fp.Write([]byte("echo(PHP_EOL);\r\n"))
	return ii, e
}

func deleteFile(fp *os.File, initialString string) (*os.File, error) {
	defer debug.SetGCPercent(100)
	defer runtime.GC()
	defer debug.FreeOSMemory()
	var err error
	fp.Truncate(0)
	fp.Seek(0, 0)
	_, err = fp.WriteAt([]byte(initialString), 0)
	fp.Seek(0, 0)
	return fp, err
}
