package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/ent/notification"
	"gitlab.smartbet.am/golang/notification/ent/partnerconfig"
	"gitlab.smartbet.am/golang/notification/ent/schema"
	"gitlab.smartbet.am/golang/notification/types"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting safe test data seeding...")

	// Database connection
	host := getEnv("DB_HOST", "localhost")
	port := "3306"
	user := getEnv("DB_USER", "notification_user")
	password := getEnv("DB_PASSWORD", "notification_pass")
	dbname := getEnv("DB_NAME", "notification_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	drv := entsql.OpenDB(dialect.MySQL, db)
	client := ent.NewClient(ent.Driver(drv))
	defer client.Close()

	ctx := context.Background()

	// Clear existing data safely
	logger.Info("Clearing existing test data...")
	clearData(ctx, client, logger)

	// Seed partner configurations
	logger.Info("Creating partner configurations...")
	seedPartnerConfigsSafe(ctx, client, logger)

	// Seed notifications
	logger.Info("Creating test notifications...")
	seedNotificationsSafe(ctx, client, logger)

	logger.Info("Safe test data seeding completed successfully!")
}

func clearData(ctx context.Context, client *ent.Client, logger *logrus.Logger) {
	// Delete in correct order (foreign key dependencies)
	deletedNotifs, _ := client.Notification.Delete().Exec(ctx)
	logger.Infof("Deleted %d existing notifications", deletedNotifs)

	deletedConfigs, _ := client.PartnerConfig.Delete().Exec(ctx)
	logger.Infof("Deleted %d existing partner configs", deletedConfigs)
}

func seedPartnerConfigsSafe(ctx context.Context, client *ent.Client, logger *logrus.Logger) {
	configs := []struct {
		id          string
		tenantID    int64
		name        string
		emailConfig string
		smsConfig   string
		pushConfig  string
		batchConfig string
		enabled     bool
	}{
		{
			id:       "goodwin-casino-1001",
			tenantID: 1001,
			name:     "Goodwin Casino",
			emailConfig: `[{
				"name": "primary",
				"type": "smtp", 
				"priority": 1,
				"enabled": true,
				"config": {
					"Host": "smtp.sendgrid.net",
					"Port": "465",
					"Username": "apikey",
					"Password": "SG.test_key_12345",
					"SMTPAuth": "1",
					"SMTPSecure": "ssl",
					"MSGBonusFrom": "bonus@goodwin.am",
					"MSGPromoFrom": "promo@goodwin.am",
					"MSGReportFrom": "reports@goodwin.am",
					"MSGSystemFrom": "noreply@goodwin.am",
					"MSGPaymentFrom": "payments@goodwin.am",
					"MSGSupportFrom": "support@goodwin.am",
					"MSGBonusFromName": "Goodwin Bonus Team",
					"MSGPromoFromName": "Goodwin Promotions",
					"MSGReportFromName": "Goodwin Reports",
					"MSGSystemFromName": "Goodwin System",
					"MSGPaymentFromName": "Goodwin Payments",
					"MSGSupportFromName": "Goodwin Support"
				}
			}]`,
			smsConfig: `[{
				"name": "primary",
				"type": "twilio",
				"priority": 1,
				"enabled": true,
				"config": {
					"account_sid": "AC_test_123456789",
					"auth_token": "test_token_123",
					"from_number": "+1234567890"
				}
			}]`,
			pushConfig: `[{
				"name": "primary",
				"type": "fcm",
				"priority": 1,
				"enabled": true,
				"config": {
					"server_key": "fcm_server_key_123",
					"project_id": "goodwin-casino"
				}
			}]`,
			batchConfig: `{
				"enabled": true,
				"max_batch_size": 100,
				"flush_interval_seconds": 10
			}`,
			enabled: true,
		},
		{
			id:       "starbet-1002",
			tenantID: 1002,
			name:     "StarBet",
			emailConfig: `[{
				"name": "primary",
				"type": "smtp",
				"priority": 1,
				"enabled": true,
				"config": {
					"Host": "smtp.sendx.io",
					"Port": "587",
					"Username": "starbet@sendx.io",
					"Password": "sendx_password_456",
					"SMTPAuth": "1",
					"SMTPSecure": "tls",
					"MSGBonusFrom": "bonuses@starbet.com",
					"MSGPromoFrom": "promotions@starbet.com",
					"MSGSystemFrom": "system@starbet.com",
					"MSGBonusFromName": "StarBet Bonuses",
					"MSGPromoFromName": "StarBet Promos",
					"MSGSystemFromName": "StarBet System"
				}
			}]`,
			smsConfig:  `[]`,
			pushConfig: `[]`,
			batchConfig: `{
				"enabled": true,
				"max_batch_size": 50,
				"flush_interval_seconds": 5
			}`,
			enabled: true,
		},
		{
			id:          "luckyplay-1003",
			tenantID:    1003,
			name:        "LuckyPlay",
			emailConfig: `[]`,
			smsConfig:   `[]`,
			pushConfig:  `[]`,
			batchConfig: `{"enabled": false}`,
			enabled:     false,
		},
	}

	rateLimits := `{
		"email": {"limit": 1000, "window": "1h", "strategy": "sliding"},
		"sms": {"limit": 500, "window": "1h", "strategy": "sliding"}, 
		"push": {"limit": 5000, "window": "1h", "strategy": "sliding"}
	}`

	for _, config := range configs {
		// Check if config exists
		exists, _ := client.PartnerConfig.Query().
			Where(partnerconfig.TenantID(config.tenantID)).
			Exist(ctx)

		if exists {
			logger.Infof("Partner config for tenant %d already exists, skipping", config.tenantID)
			continue
		}

		err := client.PartnerConfig.Create().
			SetID(config.id).
			SetTenantID(config.tenantID).
			SetEmailProviders([]byte(config.emailConfig)).
			SetSmsProviders([]byte(config.smsConfig)).
			SetPushProviders([]byte(config.pushConfig)).
			SetBatchConfig([]byte(config.batchConfig)).
			SetRateLimits([]byte(rateLimits)).
			SetEnabled(config.enabled).
			Exec(ctx)

		if err != nil {
			logger.Errorf("Failed to create partner config for %s: %v", config.name, err)
		} else {
			logger.Infof("Created partner config for %s (tenant %d)", config.name, config.tenantID)
		}
	}
}

func seedNotificationsSafe(ctx context.Context, client *ent.Client, logger *logrus.Logger) {
	// Seed notifications for existing tenants only
	tenants := []int64{1001, 1002}

	for _, tenantID := range tenants {
		// Check if tenant config exists
		exists, _ := client.PartnerConfig.Query().
			Where(partnerconfig.TenantID(tenantID)).
			Exist(ctx)

		if !exists {
			logger.Warnf("Tenant %d config not found, skipping notifications", tenantID)
			continue
		}

		createNotificationsForTenant(ctx, client, tenantID, logger)
	}
}

func createNotificationsForTenant(ctx context.Context, client *ent.Client, tenantID int64, logger *logrus.Logger) {
	now := time.Now()

	// Create different types of notifications
	notifications := []struct {
		count      int
		notifType  notification.Type
		status     notification.Status
		subject    string
		bodyPrefix string
		from       string
		scheduled  bool
	}{
		{15, notification.TypeEMAIL, notification.StatusCOMPLETED, "Welcome Bonus", "Welcome to our platform! Notification", "bonus@goodwin.am", false},
		{5, notification.TypeSMS, notification.StatusPENDING, "", "Your verification code:", "", false},
		{3, notification.TypeEMAIL, notification.StatusFAILED, "Account Update", "Important account information", "system@goodwin.am", false},
		{2, notification.TypeEMAIL, notification.StatusPENDING, "Weekend Promotion", "Don't miss our weekend special!", "promo@goodwin.am", true},
	}

	totalCreated := 0
	for _, notif := range notifications {
		for i := 0; i < notif.count; i++ {
			requestID := uuid.New().String()
			address := fmt.Sprintf("user%d@example.com", totalCreated+1)
			if notif.notifType == notification.TypeSMS {
				address = fmt.Sprintf("+1555%07d", totalCreated+1)
			}

			create := client.Notification.Create().
				SetRequestID(requestID).
				SetTenantID(tenantID).
				SetType(notif.notifType).
				SetBody(fmt.Sprintf("%s #%d", notif.bodyPrefix, i+1)).
				SetAddress(types.Address(address)).
				SetStatus(notif.status)

			if notif.subject != "" {
				create.SetHeadline(notif.subject)
			}
			if notif.from != "" {
				create.SetFrom(notif.from)
			}
			if notif.scheduled {
				futureTime := now.Add(24 * time.Hour).Unix()
				create.SetScheduleTs(futureTime)
			}
			if notif.status == notification.StatusFAILED {
				create.SetErrorMessage("Delivery failed - invalid address format")
				create.SetRetryCount(2)
			}

			// Add metadata
			create.SetMeta(&schema.NotificationMeta{
				Service: "test-service",
				Params: map[string]interface{}{
					"message_type": "system",
					"tenant_name":  fmt.Sprintf("tenant_%d", tenantID),
				},
			})

			if err := create.Exec(ctx); err != nil {
				logger.Errorf("Failed to create notification: %v", err)
			} else {
				totalCreated++
			}
		}
	}

	logger.Infof("Created %d notifications for tenant %d", totalCreated, tenantID)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
