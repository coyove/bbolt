package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"syscall/js"
	"time"

	"github.com/coyove/bbolt"
)

var db *bbolt.DB

func consoleError(msg string, args ...any) {
	now := time.Now()
	js.Global().
		Get("console").
		Call("error", js.ValueOf(fmt.Sprintf(now.Format("[BBOLT 15:04:05] ")+msg, args...)))
}

func main() {
	js.Global().Set("bbOpen", js.FuncOf(func(this js.Value, args []js.Value) any {
		var err error
		var data []byte
		if len(args) > 0 && args[0].Length() > 0 {
			data = make([]byte, args[0].Length())
			js.CopyBytesToGo(data, args[0])
		}
		db, err = bbolt.Open("<memory>", 0777, &bbolt.Options{
			FreelistType: bbolt.FreelistMapType,
			MemData:      data,
		})
		if err != nil {
			consoleError("failed to open data: %v", err)
			return nil
		}
		db.BeforeCommit = func(patch []byte) error {
			u8 := js.Global().Get("Uint8Array").New(len(patch))
			js.CopyBytesToJS(u8, patch)
			res := args[1].Invoke(u8)
			if res.Type() == js.TypeString {
				return errors.New(res.String())
			}
			return nil
		}
		return nil
	}))

	js.Global().Set("bbSet", js.FuncOf(func(this js.Value, args []js.Value) any {
		err := db.Update(func(tx *bbolt.Tx) error {
			db, _ := tx.CreateBucketIfNotExists([]byte(args[0].String()))
			return db.Put([]byte(args[1].String()), []byte(args[2].String()))
		})
		if err != nil {
			consoleError("failed to set data: %v", err)
			return nil
		}
		return nil
	}))

	js.Global().Set("bbDump", js.FuncOf(func(this js.Value, args []js.Value) any {
		out := &bytes.Buffer{}
		w := gzip.NewWriter(out)
		w.Write(db.UnsafeMemData())
		w.Close()
		u8 := js.Global().Get("Uint8Array").New(out.Len())
		js.CopyBytesToJS(u8, out.Bytes())
		return u8
	}))

	select {}
}
