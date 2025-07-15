package main

import (
	"context"
	"database/sql"
	"encoding/json"
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
		id             string
		tenantID       int64
		name           string
		emailProviders []schema.ProviderConfig
		smsProviders   []schema.ProviderConfig
		pushProviders  []schema.ProviderConfig
		batchConfig    *schema.BatchConfig
		rateLimits     map[string]schema.RateLimit
		enabled        bool
	}{
		{
			id:       "goodwin-casino-1001",
			tenantID: 1001,
			name:     "Goodwin Casino",
			emailProviders: []schema.ProviderConfig{
				{
					Name:     "primary",
					Type:     "smtp",
					Priority: 1,
					Enabled:  true,
					Config: map[string]interface{}{
						"Host":               "smtp.sendgrid.net",
						"Port":               "465",
						"Username":           "apikey",
						"Password":           "SG.test_key_12345",
						"SMTPAuth":           "1",
						"SMTPSecure":         "ssl",
						"MSGBonusFrom":       "bonus@goodwin.am",
						"MSGPromoFrom":       "promo@goodwin.am",
						"MSGReportFrom":      "reports@goodwin.am",
						"MSGSystemFrom":      "noreply@goodwin.am",
						"MSGPaymentFrom":     "payments@goodwin.am",
						"MSGSupportFrom":     "support@goodwin.am",
						"MSGBonusFromName":   "Goodwin Bonus Team",
						"MSGPromoFromName":   "Goodwin Promotions",
						"MSGReportFromName":  "Goodwin Reports",
						"MSGSystemFromName":  "Goodwin System",
						"MSGPaymentFromName": "Goodwin Payments",
						"MSGSupportFromName": "Goodwin Support",
					},
				},
			},
			smsProviders: []schema.ProviderConfig{
				{
					Name:     "primary",
					Type:     "twilio",
					Priority: 1,
					Enabled:  true,
					Config: map[string]interface{}{
						"account_sid": "AC_test_123456789",
						"auth_token":  "test_token_123",
						"from_number": "+1234567890",
					},
				},
			},
			pushProviders: []schema.ProviderConfig{
				{
					Name:     "primary",
					Type:     "fcm",
					Priority: 1,
					Enabled:  true,
					Config: map[string]interface{}{
						"server_key": "fcm_server_key_123",
						"project_id": "goodwin-casino",
					},
				},
			},
			batchConfig: &schema.BatchConfig{
				Enabled:              true,
				MaxBatchSize:         100,
				FlushIntervalSeconds: 10,
			},
			rateLimits: map[string]schema.RateLimit{
				"email": {Limit: 1000, Window: "1h", Strategy: "sliding"},
				"sms":   {Limit: 500, Window: "1h", Strategy: "sliding"},
				"push":  {Limit: 5000, Window: "1h", Strategy: "sliding"},
			},
			enabled: true,
		},
		{
			id:       "starbet-1002",
			tenantID: 1002,
			name:     "StarBet",
			emailProviders: []schema.ProviderConfig{
				{
					Name:     "primary",
					Type:     "smtp",
					Priority: 1,
					Enabled:  true,
					Config: map[string]interface{}{
						"Host":              "smtp.sendx.io",
						"Port":              "587",
						"Username":          "starbet@sendx.io",
						"Password":          "sendx_password_456",
						"SMTPAuth":          "1",
						"SMTPSecure":        "tls",
						"MSGBonusFrom":      "bonuses@starbet.com",
						"MSGPromoFrom":      "promotions@starbet.com",
						"MSGSystemFrom":     "system@starbet.com",
						"MSGBonusFromName":  "StarBet Bonuses",
						"MSGPromoFromName":  "StarBet Promos",
						"MSGSystemFromName": "StarBet System",
					},
				},
			},
			smsProviders:  []schema.ProviderConfig{},
			pushProviders: []schema.ProviderConfig{},
			batchConfig: &schema.BatchConfig{
				Enabled:              true,
				MaxBatchSize:         50,
				FlushIntervalSeconds: 5,
			},
			rateLimits: map[string]schema.RateLimit{
				"email": {Limit: 500, Window: "1h", Strategy: "sliding"},
				"sms":   {Limit: 100, Window: "1h", Strategy: "sliding"},
				"push":  {Limit: 1000, Window: "1h", Strategy: "sliding"},
			},
			enabled: true,
		},
		{
			id:             "luckyplay-1003",
			tenantID:       1003,
			name:           "LuckyPlay",
			emailProviders: []schema.ProviderConfig{},
			smsProviders:   []schema.ProviderConfig{},
			pushProviders:  []schema.ProviderConfig{},
			batchConfig: &schema.BatchConfig{
				Enabled: false,
			},
			rateLimits: map[string]schema.RateLimit{
				"email": {Limit: 100, Window: "1h", Strategy: "sliding"},
				"sms":   {Limit: 50, Window: "1h", Strategy: "sliding"},
				"push":  {Limit: 200, Window: "1h", Strategy: "sliding"},
			},
			enabled: false,
		},
	}

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
			SetEmailProviders(config.emailProviders).
			SetSmsProviders(config.smsProviders).
			SetPushProviders(config.pushProviders).
			SetBatchConfig(config.batchConfig).
			SetRateLimits(config.rateLimits).
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
			metaData := map[string]interface{}{
				"service": "test-service",
				"params": map[string]interface{}{
					"message_type": "system",
					"tenant_name":  fmt.Sprintf("tenant_%d", tenantID),
				},
			}

			metaJSON, _ := json.Marshal(metaData)
			var meta schema.NotificationMeta
			json.Unmarshal(metaJSON, &meta)
			create.SetMeta(&meta)

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
