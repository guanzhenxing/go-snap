package logger

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// 预编译的正则表达式模式，减少运行时开销
var (
	// 信用卡号：匹配16位数字，可能有分隔符
	creditCardRegexPattern = regexp.MustCompile(`\b(?:\d{4}[- ]?){3}\d{4}\b`)

	// 中国手机号：匹配1开头的11位数字
	phoneRegexPattern = regexp.MustCompile(`\b1[3-9]\d{9}\b`)

	// 电子邮件：RFC 5322规范的简化版
	emailRegexPattern = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)

	// 中国身份证号：18位，最后一位可能是X
	idCardRegexPattern = regexp.MustCompile(`\b\d{17}[\dXx]\b`)

	// JSON字段正则表达式缓存
	jsonRegexCache   = make(map[string]*regexp.Regexp)
	jsonRegexCacheMu sync.RWMutex
)

// MaskField 通用脱敏处理
func MaskField(key string, value interface{}, maskChar string) Field {
	if value == nil {
		return String(key, "")
	}

	strValue := fmt.Sprintf("%v", value)
	if len(strValue) == 0 {
		return String(key, "")
	}

	// 短字符串全部脱敏
	if len(strValue) <= 6 {
		return String(key, strings.Repeat(maskChar, len(strValue)))
	}

	// 保留前三后三，中间脱敏
	masked := strValue[:3] + strings.Repeat(maskChar, len(strValue)-6) + strValue[len(strValue)-3:]
	return String(key, masked)
}

// getJSONFieldRegex 获取JSON字段的正则表达式（带缓存）
func getJSONFieldRegex(key string) *regexp.Regexp {
	// 先检查缓存
	jsonRegexCacheMu.RLock()
	re, exists := jsonRegexCache[key]
	jsonRegexCacheMu.RUnlock()

	if exists {
		return re
	}

	// 缓存未命中，创建新的正则表达式
	pattern := fmt.Sprintf(`("%s"\s*:\s*")(.*?)("(?:\s*,|\s*}))`, regexp.QuoteMeta(key))
	re = regexp.MustCompile(pattern)

	// 存入缓存
	jsonRegexCacheMu.Lock()
	jsonRegexCache[key] = re
	jsonRegexCacheMu.Unlock()

	return re
}

// MaskJSON 对JSON字符串中的敏感字段进行脱敏
func MaskJSON(json string, sensitiveKeys []string) string {
	if len(json) == 0 || len(sensitiveKeys) == 0 {
		return json
	}

	result := json
	for _, key := range sensitiveKeys {
		re := getJSONFieldRegex(key)
		result = re.ReplaceAllStringFunc(result, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 4 {
				value := parts[2]
				if len(value) <= 6 {
					maskedValue := strings.Repeat("*", len(value))
					return parts[1] + maskedValue + parts[3]
				}
				maskedValue := value[:3] + strings.Repeat("*", len(value)-6) + value[len(value)-3:]
				return parts[1] + maskedValue + parts[3]
			}
			return match
		})
	}
	return result
}

// MaskCreditCard 信用卡号脱敏
func MaskCreditCard(key string, value string) Field {
	if value == "" {
		return String(key, "")
	}

	// 移除空格和破折号
	cleanValue := strings.ReplaceAll(strings.ReplaceAll(value, " ", ""), "-", "")
	if len(cleanValue) < 13 || len(cleanValue) > 19 {
		return String(key, value) // 不符合信用卡号长度，不处理
	}

	// 保留前4位和后4位
	masked := cleanValue[:4] + strings.Repeat("*", len(cleanValue)-8) + cleanValue[len(cleanValue)-4:]
	return String(key, masked)
}

// MaskPhone 电话号码脱敏
func MaskPhone(key string, value string) Field {
	if value == "" {
		return String(key, "")
	}

	// 中国手机号处理
	cleanValue := strings.ReplaceAll(strings.ReplaceAll(value, " ", ""), "-", "")
	if len(cleanValue) != 11 {
		return MaskField(key, value, "*") // 不是标准手机号，使用通用脱敏
	}

	// 保留前三位和后四位
	masked := cleanValue[:3] + "****" + cleanValue[len(cleanValue)-4:]
	return String(key, masked)
}

// MaskEmail 电子邮件脱敏
func MaskEmail(key string, value string) Field {
	if value == "" || !strings.Contains(value, "@") {
		return String(key, value)
	}

	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return MaskField(key, value, "*")
	}

	username := parts[0]
	domain := parts[1]

	var maskedUsername string
	if len(username) <= 3 {
		maskedUsername = username[:1] + strings.Repeat("*", len(username)-1)
	} else {
		maskedUsername = username[:3] + strings.Repeat("*", len(username)-3)
	}

	return String(key, maskedUsername+"@"+domain)
}

// MaskIDCard 身份证号脱敏
func MaskIDCard(key string, value string) Field {
	if value == "" {
		return String(key, "")
	}

	cleanValue := strings.ReplaceAll(value, " ", "")
	if len(cleanValue) != 18 {
		return MaskField(key, value, "*") // 不是标准身份证号，使用通用脱敏
	}

	// 保留前6位和后4位，中间脱敏
	masked := cleanValue[:6] + strings.Repeat("*", 8) + cleanValue[len(cleanValue)-4:]
	return String(key, masked)
}

// AutoMask 根据内容自动选择脱敏策略
func AutoMask(key string, value string) Field {
	if value == "" {
		return String(key, "")
	}

	// 根据内容特征选择脱敏策略
	if creditCardRegexPattern.MatchString(value) {
		return MaskCreditCard(key, value)
	}

	if phoneRegexPattern.MatchString(value) {
		return MaskPhone(key, value)
	}

	if emailRegexPattern.MatchString(value) {
		return MaskEmail(key, value)
	}

	if idCardRegexPattern.MatchString(value) {
		return MaskIDCard(key, value)
	}

	// 默认脱敏处理
	return MaskField(key, value, "*")
}
