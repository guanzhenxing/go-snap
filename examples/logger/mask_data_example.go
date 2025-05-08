package main

import (
	"github.com/guanzhenxing/go-snap/logger"
)

func mainMaskData() {
	// 初始化日志
	logger.Init(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithJSONConsole(true),
		// 设置自动掩码的敏感字段
		logger.WithSensitiveKeys("password", "credit_card", "phone", "email", "id_card"),
	)

	// 记录日志时自动掩码敏感信息
	logger.Info("用户注册成功",
		logger.String("username", "zhangsan"),
		logger.String("password", "123456"),              // 敏感信息会被掩码
		logger.String("email", "zhangsan@example.com"),   // 敏感信息会被掩码
		logger.String("phone", "13800138000"),            // 敏感信息会被掩码
		logger.String("credit_card", "6225881234567890"), // 敏感信息会被掩码
		logger.String("id_card", "110101199001011234"),   // 敏感信息会被掩码
		logger.String("address", "北京市海淀区中关村"),
	)

	// 手动掩码特定字段
	email := "zhangsan@example.com"
	logger.Info("掩码邮箱示例",
		logger.MaskEmail("email", email),
	)

	// 掩码信用卡
	creditCard := "6225881234567890"
	logger.Info("掩码信用卡示例",
		logger.MaskCreditCard("credit_card", creditCard),
	)

	// 掩码手机号
	phone := "13800138000"
	logger.Info("掩码手机号示例",
		logger.MaskPhone("phone", phone),
	)

	// 掩码身份证
	idCard := "110101199001011234"
	logger.Info("掩码身份证示例",
		logger.MaskIDCard("id_card", idCard),
	)

	// 掩码JSON数据中的字段
	jsonData := `{
		"user": {
			"name": "张三",
			"email": "zhangsan@example.com",
			"password": "123456",
			"card": "6225881234567890",
			"details": {
				"id_card": "110101199001011234",
				"phone": "13800138000"
			}
		}
	}`

	// 掩码JSON数据中的敏感字段
	maskedJSON := logger.MaskJSON(jsonData, []string{"email", "password", "card", "id_card", "phone"})
	logger.Info("掩码后的JSON数据", logger.String("user_data", maskedJSON))

	// 自定义字段掩码
	customData := "我的密码是123456，请保密"
	logger.Info("自定义掩码示例",
		logger.MaskField("data", customData, "*"),
	)

	// 确保日志刷新
	logger.Sync()
}
