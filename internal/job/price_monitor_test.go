package job

import (
	"klineio/internal/service"
	"testing"
)

func TestPriceMonitorJob_Type(t *testing.T) {
	// 只做类型实例化测试，保证测试文件有内容且能编译
	var svc *service.PriceMonitorService
	job := &PriceMonitorJob{
		priceMonitorSvc: svc,
		logger:          nil,
	}
	if job.priceMonitorSvc != nil {
		t.Errorf("expected nil service")
	}
}
