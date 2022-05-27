// Created on 2022/5/27.
// @author tony
// email xmgtony@gmail.com
// description 账目清单handler

package accountbill

import (
	"apiserver-gin/internal/model"
	"apiserver-gin/internal/service"
	"apiserver-gin/pkg/constant"
	"apiserver-gin/pkg/errors"
	"apiserver-gin/pkg/errors/ecode"
	"apiserver-gin/pkg/response"
	jtime "apiserver-gin/pkg/time"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

// Handler 账目清单handler，从分层来讲，这里已经是最外层，只要提供实例在router中使用
// 所以这里没有定义接口，而是直接使用struct来组织多个handler func
type Handler struct {
	accountBillServ service.AccountBillService
}

func NewAccountBillHandler(_accountBillServ service.AccountBillService) *Handler {
	return &Handler{
		accountBillServ: _accountBillServ,
	}
}

func (abh *Handler) AddAccountBill() gin.HandlerFunc {
	return func(c *gin.Context) {
		addAccountBillReq := model.AddAccountBillReq{}
		if err := c.ShouldBind(&addAccountBillReq); err != nil {
			response.JSON(c, errors.WithCode(ecode.ValidateErr, err.Error()), nil)
			return
		}
		uid := c.GetInt64(constant.UserID)
		amd, err := decimal.NewFromString(addAccountBillReq.Amount)
		if err != nil {
			response.JSON(c, errors.Wrap(err, ecode.ValidateErr, "金额必须为有效数字"), nil)
			return
		}
		if amd.IsNegative() {
			response.JSON(c, errors.Wrap(err, ecode.ValidateErr, "金额必须为正数"), nil)
			return
		}
		strings.SplitN(addAccountBillReq.Amount, ",", 2)
		amount := amd.Mul(decimal.NewFromInt32(100)).IntPart()
		// 组织model信息
		accountBill := model.AccountBill{
			UserId:         uint64(uid),
			BillDate:       time.Time(addAccountBillReq.BillDate),
			OriginIncident: addAccountBillReq.OriginIncident,
			Amount:         uint(amount),
			Relation:       addAccountBillReq.Relation,
			ToName:         addAccountBillReq.ToName,
			IsFollow:       addAccountBillReq.IsFollow,
			Remark:         addAccountBillReq.Remark,
		}
		err = abh.accountBillServ.Save(c, &accountBill)
		if err != nil {
			response.JSON(c, errors.Wrap(err, ecode.RecordCreateErr, "保存账目清单信息失败"), nil)
			return
		}
		response.JSON(c, nil, nil)
	}
}

func (abh *Handler) GetAccountBillList() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用auth中间件的接口，一定能解析出来，否则中间件会响应错误
		uid := c.GetInt64(constant.UserID)
		bills, err := abh.accountBillServ.SelectListByUserId(c, uid)
		if err != nil {
			response.JSON(c, errors.Wrap(err, ecode.NotFoundErr, "查询错误，未找到记录"), nil)
			return
		}
		respBills := make([]model.AccountBillResp, 0)
		for _, bill := range bills {
			// 使用BigDecimal做精确运算，避免丢失精度
			amountStr := decimal.NewFromInt32(int32(bill.Amount)).
				Div(decimal.NewFromInt32(100)).StringFixed(2)

			respBill := model.AccountBillResp{
				BillDate:       jtime.JsonTime(bill.BillDate),
				OriginIncident: bill.OriginIncident,
				Amount:         amountStr,
				Relation:       bill.Relation,
				ToName:         bill.ToName,
				IsFollow:       bill.IsFollow,
				Remark:         bill.Remark,
			}
			respBills = append(respBills, respBill)
		}
		response.JSON(c, nil, respBills)
		return
	}
}
