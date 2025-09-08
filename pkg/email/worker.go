package email

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/perfect-panel/server/internal/model/task"
	"github.com/perfect-panel/server/pkg/logger"
	"gorm.io/gorm"
)

type ErrorInfo struct {
	Error string `json:"error"`
	Email string `json:"email"`
	Time  int64  `json:"time"`
}

type Worker struct {
	id     int64           // 任务ID
	db     *gorm.DB        // 数据库连接
	ctx    context.Context // 上下文
	sender Sender          // 邮件发送器接口
	status uint8           // 任务状态，0 表示未运行，1 表示运行中 2 表示已完成
}

func NewWorker(ctx context.Context, id int64, db *gorm.DB, sender Sender) *Worker {
	return &Worker{
		id:     id,
		db:     db,
		ctx:    ctx,
		sender: sender,
	}
}

// GetID 获取Worker的任务ID
func (w *Worker) GetID() int64 {
	return w.id
}

// IsRunning 检查Worker是否正在运行
func (w *Worker) IsRunning() uint8 {
	return w.status
}

// Start 启动Worker，开始处理任务
func (w *Worker) Start() {
	// 检查并发限制
	limit.Lock()
	defer limit.Unlock()
	tx := w.db.WithContext(w.ctx)
	var taskInfo task.EmailTask
	if err := tx.Model(&task.EmailTask{}).Where("id = ?", w.id).First(&taskInfo).Error; err != nil {
		logger.Error("Batch Send Email",
			logger.Field("message", "Failed to find task"),
			logger.Field("error", err.Error()),
			logger.Field("task_id", w.id),
		)
		w.status = 2 // 设置状态为已完成
		return
	}
	if taskInfo.Status != 0 {
		logger.Error("Batch Send Email",
			logger.Field("message", "Task already completed or in progress"),
			logger.Field("task_id", w.id),
		)
		w.status = 2 // 设置状态为已完成
		return
	}
	if taskInfo.Recipients == "" && taskInfo.Additional == "" {
		logger.Error("Batch Send Email",
			logger.Field("message", "No recipients or additional emails provided"),
			logger.Field("task_id", w.id),
		)
		w.status = 2 // 设置状态为已完成
		return
	}
	w.status = 1 // 设置状态为运行中
	var recipients []string
	// 解析收件人
	if taskInfo.Recipients != "" {
		recipients = append(recipients, strings.Split(taskInfo.Recipients, "\n")...)
	}
	// 解析附加收件人
	if taskInfo.Additional != "" {
		recipients = append(recipients, strings.Split(taskInfo.Additional, "\n")...)
	}
	if len(recipients) == 0 {
		logger.Error("Batch Send Email",
			logger.Field("message", "No valid recipients found"),
			logger.Field("task_id", w.id),
		)
		w.status = 2 // 设置状态为已完成
		return
	}

	// 设置发送间隔时间
	var intervalTime time.Duration
	if taskInfo.Interval == 0 {
		intervalTime = 1 * time.Second
	} else {
		intervalTime = time.Duration(taskInfo.Interval) * time.Second
	}

	var errors []ErrorInfo
	var count uint64
	for _, recipient := range recipients {
		select {
		case <-w.ctx.Done():
			logger.Info("Batch Send Email",
				logger.Field("message", "Worker stopped by context cancellation"),
				logger.Field("task_id", w.id),
			)
			return
		default:
		}

		if taskInfo.Status == 0 {
			taskInfo.Status = 1 // 1 表示任务进行中
		}

		if err := w.sender.Send([]string{recipient}, taskInfo.Subject, taskInfo.Content); err != nil {
			logger.Error("Batch Send Email",
				logger.Field("message", "Failed to send email"),
				logger.Field("error", err.Error()),
				logger.Field("recipient", recipient),
				logger.Field("task_id", w.id),
			)
			errors = append(errors, ErrorInfo{
				Error: err.Error(),
				Email: recipient,
				Time:  time.Now().Unix(),
			})
			text, _ := json.Marshal(errors)
			taskInfo.Errors = string(text)
		}
		count++
		taskInfo.Current = count
		if err := tx.Model(&task.EmailTask{}).Where("`id` = ?", taskInfo.Id).Save(&taskInfo).Error; err != nil {
			logger.Error("Batch Send Email",
				logger.Field("message", "Failed to update task progress"),
				logger.Field("error", err.Error()),
				logger.Field("task_id", w.id),
			)
			w.status = 2 // 设置状态为已完成
			return
		}
		time.Sleep(intervalTime)
	}
	taskInfo.Status = 2 // 设置状态为已完成
	if err := tx.Model(&task.EmailTask{}).Where("`id` = ?", taskInfo.Id).Save(&taskInfo).Error; err != nil {
		logger.Error("Batch Send Email",
			logger.Field("message", "Failed to finalize task"),
			logger.Field("error", err.Error()),
			logger.Field("task_id", w.id),
		)
	} else {
		logger.Info("Batch Send Email",
			logger.Field("message", "Task completed successfully"),
			logger.Field("task_id", w.id),
			logger.Field("total_sent", count),
		)
	}
}
