package retry

import (
	"fmt"
	"time"

	"github.com/AleGaliev/runtimemetrics/internal/service/checkerror"
)

type Retry struct {
	RetryCount int
	Interval   []time.Duration
}

func CreateRetry() Retry {
	return Retry{
		RetryCount: 3,
		Interval:   []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

func (retry *Retry) RetryConnection(dbFunction func() error) error {
	var err error
	for i := 0; i < retry.RetryCount; i++ {

		err = dbFunction()
		if err == nil {
			return nil
		}

		if !checkerror.IsRetriableError(err) {
			return fmt.Errorf("connection problem: %v", err)
		}
		fmt.Printf("retrying attempt %d/%d\n", i+1, retry.RetryCount)
		time.Sleep(retry.Interval[i])
	}

	return fmt.Errorf("max retry connect: %v", err)
}
