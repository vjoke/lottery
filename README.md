# lottery

彩票
该智能合约实现了一个简单的彩票系统。 由于条约公开无法随意篡改的特性， 本彩票系统天然具备稳定性和中立性。

该智能合约中三种角色如下：

运营商 Company，负责制定规则。
彩票发行人 Issuer，负责发行彩票
投注人 Player


发行人发起交易，将彩票类型、彩票底金、投注时间、彩票钱包地址及公钥、发行人钱包地址、运营商钱包地址等数据写入区块，生成第一张彩票即奖池。投注人投注导入彩票池信息，然后与发行人写入区块中的彩票信息进行核对，如果彩票有效，将发起交易，即投注，投注成功后，也会将自己的钱包地址、投注号码、投注金额等数据写入区块。生成自己的彩票即投注券。等到开奖日期，中奖号码产生，这之后奖券持有人导入彩票信息与区块核对，如果有效即可领奖，领奖成功后，写入区块中注明。至此，领奖成功。

通过上述过程不难发现，“彩票”发行的每一个环节，都被写入到了区块链中，防止恶意的作弊行为。发行人将彩票发行信息写入到区块中，彩票投注者将投注信息写到区块链中，中奖的人也将自己兑奖信息写入到区块中，这就形成一个完整的闭环系统，三者之间互相制约。在区块链这本大帐单中，保留了彩票交易每个细节，被全网用户所共同享有，任何单方面的作弊行为，都会被网络拒绝。

账户私钥应该由安装在本地的客户端生成，本合约中为了简便，使用模拟私钥和公钥。


规则：
	1. 彩票发行时初始注入资金作为奖池，不得低于系统设定的最低值。
	2. 开奖日期超出预设一天，扣除相应罚金, 关闭彩票日期顺延。
	3. 关闭彩票时系统按收益提取一定比例的费用。
	4. 单注彩票中奖金额 = 奖池金额/所有中奖彩票数。
	5. 逾期未兑奖不予处理
	

数据结构设计

运营公司
	名称
	钱包地址
	公钥
	私钥

发行人
	名称
	钱包地址	
	公钥
	私钥

彩票
	名称
	地址
	类型
	底金
	开奖日期
	截止日期
	彩票钱包地址
	公钥
	私钥
	所有投注券地址
	发行人钱包地址
	中奖号码
	状态
	单注中奖金额
	运营商钱包地址

投注券
	地址
	彩票地址
	玩家钱包地址
	玩家签名
	投注号码
	投注票数
	状态

兑奖
	同投注券，状态为已兑奖

对彩票的所有操作都归为记录。


function及各自实现的功能
	init 初始化函数
	invoke 调用合约内部的函数
	query 查询相关的信息

	createCompany 创建运营公司
	createIssuer 创建彩票发行方
	createPlayer 创建玩家
	
	createLottery 发行彩票
	drawLottery 开奖，开始兑奖
	closeLottery 截止日期到，关闭彩票

	buyTicket 购买彩票
	takePrize 兑奖

