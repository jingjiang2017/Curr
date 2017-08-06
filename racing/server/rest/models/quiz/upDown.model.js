const mongoose = require('mongoose');
const Schema = mongoose.Schema;

//竞猜记录
const UpDownSchema = new Schema({

    username: {type: String, required: true}, // 用户名
    avatar: {type: String, default: ''},    // 头像
    balance:  {type: Number, default: 0}, //当前余额
    type: {type: Boolean, default: true},  //上分: true /下分: false
    amount: {type: Number, default: 0}, // 上下分金额
    backMethod: {type:String}, //提现方式
    backNo: {type: String}, //提现账号
    byWho: {type: String}, //操作人员
    profile: {type: String}, //备注
    createdAt: {type: Date, default: Date.now}, //申请时间
    updateAt: {type: Date}, // 审批时间
});

module.exports = mongoose.model('UpDown', UpDownSchema);