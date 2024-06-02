package errs

const (
	CodeBadRequest   = 300
	CodeUnauthorized = 401
	CodeForbidden    = 403

	CodeUserNotFound       = 1001
	CodeUserExists         = 1002
	CodeVerificationError  = 1003
	CodeBadCredentials     = 1004
	CodeCredentialsExpired = 1005

	CodeContactExists               = 2001
	CodeContactRequestExists        = 2002
	CodeContactRequestNotFound      = 2003
	CodeContactRequestStatusInvalid = 2004
	CodeContactNotFound             = 2005
	CodeContactRequestExpired       = 2006
	CodeMessageRevokeExpired        = 2007
	CodeMessageEmpty                = 2008
	CodeMessageTypeNotSupported     = 2009

	CodeCallStatusInvalid        = 3001
	CodeCallMemberCountNotEnough = 3002
	CodeCallUserLockInvalid      = 3003
	CodeCallManagerLocked        = 3004
	CodeCallManagerNotFound      = 3005
	CodeCallNotFound             = 3006
	CodeCalleeBusy               = 3007
	CodeCallerBusy               = 3008
	CodeCallStatusNotReady       = 3009

	CodeSaveFileFailed = 4001
)

var UserNotFound = NewAppError(CodeUserNotFound, "用户不存在")
var UserExists = NewAppError(CodeUserExists, "用户已存在")
var BadCredentials = NewAppError(CodeBadCredentials, "用户名或密码错误")
var Unauthorized = NewAppError(CodeUnauthorized, "未登录")
var Forbidden = NewAppError(CodeForbidden, "无权限")
var CredentialsExpired = NewAppError(CodeCredentialsExpired, "凭证已过期")

var ContactExists = NewAppError(CodeContactExists, "联系人已存在")
var ContactNotFound = NewAppError(CodeContactNotFound, "联系人不存在")
var ContactRequestExists = NewAppError(CodeContactRequestExists, "联系人申请已存在")
var ContactRequestNotFound = NewAppError(CodeContactRequestNotFound, "联系人申请不存在")
var ContactRequestStatusInvalid = NewAppError(CodeContactRequestStatusInvalid, "联系人申请状态无效")
var ContactRequestExpired = NewAppError(CodeContactRequestExpired, "联系人申请已过期")
var MessageRevokeExpired = NewAppError(CodeMessageRevokeExpired, "只能撤回两分钟内的消息")
var MessageEmpty = NewAppError(CodeMessageEmpty, "消息内容为空")
var MessageTypeNotSupported = NewAppError(CodeMessageTypeNotSupported, "不支持的消息类型")

var CallStatusInvalid = NewAppError(CodeCallStatusInvalid, "通话状态无效")
var CallStatusNotReady = NewAppError(CodeCallStatusNotReady, "通话状态未准备")
var CallNotFound = NewAppError(CodeCallNotFound, "通话不存在")
var CalleeBusy = NewAppError(CodeCalleeBusy, "被叫用户忙")
var CallerBusy = NewAppError(CodeCallerBusy, "主叫用户忙")
var CallMemberCountNotEnough = NewAppError(CodeCallMemberCountNotEnough, "通话人数不足")
var CallUserLockInvalid = NewAppError(CodeCallUserLockInvalid, "通话用户锁无效")
var CallManagerLocked = NewAppError(CodeCallManagerLocked, "通话管理锁已被占用")
var CallManagerNotFound = NewAppError(CodeCallManagerNotFound, "通话管理器不存在")

var SaveFileFailed = NewAppError(CodeSaveFileFailed, "文件保存失败")

func NewVerificationError(error string) AppError {
	return NewAppError(CodeVerificationError, error)
}
