package shared

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/xtaci/smux"
)

func PipeServer(dst net.Conn, src io.Reader, stream *smux.Stream) {
	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(dst, src)
		if cw, ok := dst.(interface{ CloseWrite() error }); ok {
			_ = cw.CloseWrite()
		} else {
			_ = dst.Close()
		}
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(stream, dst)
		stream.Close()
		done <- struct{}{}
	}()
	<-done
	<-done
}

func PipeAgent(a, b io.ReadWriteCloser) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(a, b)
		if cw, ok := a.(interface{ CloseWrite() error }); ok {
			_ = cw.CloseWrite()
		} else {
			_ = a.Close()
		}
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(b, a)
		_ = b.Close()
	}()

	wg.Wait()
	time.Sleep(50 * time.Millisecond)
}
