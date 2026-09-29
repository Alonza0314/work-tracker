package processor

import (
	"backend/internal/context"
	"backend/logger"
	"net/http"
	"time"
)

type ProcessorIE struct {
	Username string
	Password string

	JwtSecret    string
	JwtExpiresIn time.Duration

	HolidaySync      bool
	HolidaySourceUrl string

	*context.SystemContext

	*logger.BackendLogger
}

type Processor struct {
	username string
	password string

	jwtSecret    string
	jwtExpiresIn time.Duration

	holidaySync      bool
	holidaySourceUrl string
	httpClient       *http.Client
	// closed by Release to stop the background holiday sync
	holidaySyncStop chan struct{}
	holidaySyncDone chan struct{}

	// the clock; replaced in tests
	now func() time.Time

	*context.SystemContext

	*logger.BackendLogger
}

func NewProcessor(ie *ProcessorIE) *Processor {
	return &Processor{
		username: normalizeAccount(ie.Username),
		password: ie.Password,

		jwtSecret:    ie.JwtSecret,
		jwtExpiresIn: ie.JwtExpiresIn,

		holidaySync:      ie.HolidaySync,
		holidaySourceUrl: ie.HolidaySourceUrl,
		httpClient:       &http.Client{Timeout: 15 * time.Second},
		holidaySyncStop:  nil,
		holidaySyncDone:  nil,

		now: time.Now,

		SystemContext: ie.SystemContext,

		BackendLogger: ie.BackendLogger,
	}
}

func (p *Processor) Release() {
	p.ProcLog.Infoln("Release processor...")

	p.stopHolidaySync()
	p.SystemContext.Release()

	p.ProcLog.Infoln("Processor released")
}
