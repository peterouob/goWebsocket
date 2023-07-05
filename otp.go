package main

import (
	"context"
	"github.com/google/uuid"
	"time"
)

type OTP struct {
	Key     string    `json:"key"`
	Created time.Time `json:"created"`
}

// 這個map將包含一次性密碼，刪除太舊的一次性密碼
type RetentionMap map[string]OTP

// 接受一個上下文和保留時間
func NewRetentionMap(ctx context.Context, retentiontime time.Duration) RetentionMap {
	rm := make(RetentionMap)
	go rm.Retention(ctx, retentiontime) //開一個協程不斷檢查
	return rm
}

func (rm RetentionMap) NewOTP() OTP {
	o := OTP{
		Key:     uuid.NewString(),
		Created: time.Now(),
	}
	rm[o.Key] = o
	return o
}

func (rm RetentionMap) VerifyOTP(otp string) bool {
	if _, ok := rm[otp]; !ok {
		return false // otp is not valid
	}
	delete(rm, otp) //刪除一次性密碼
	return true
}

func (rm RetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	//每次重新檢查的頻率
	ticker := time.NewTicker(400 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			for _, otp := range rm {
				//過期時間比現在早，這密碼無效
				if otp.Created.Add(retentionPeriod).Before(time.Now()) {
					delete(rm, otp.Key)
				}
			}
			//上下文關閉
		case <-ctx.Done():
			return
		}
	}
}
