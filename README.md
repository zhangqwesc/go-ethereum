## Go Ethereum

在[go-ethereum](https://github.com/ethereum/go-ethereum)的基础上，增加
- getTokenTransfer 用于查询某个地址的特定erc20代币的交易记录
- getTokens 用于查询某个地址参与过交易的代币（不论是from还是to，也不论balance是否为0）

在`master`分支中，则是增加了getAddressTransactions用于获取任意地址的交易记录
### 使用方法
在`geth console`中

    eth.getTokenTransfer(address, tokenAddress, start, end)
    eth.getTokens(address, start, end)

在`web3.js`中

    const Web3 = require("web3")
    const net = require("net")
    const web3 = new Web3("/home/user/.ethereum/geth.ipc", net);

    web3.eth.extend({
      methods: [{
        name: "getTokenTransfer",
        call: 'eth_getTokenTransfer',
        params: 4,
        inputFormatter: [web3.extend.formatters.inputAddressFormatter, web3.extend.formatters.inputAddressFormatter, null, null]
      }, {
        name: "getTokens",
        call: 'eth_getTokens',
        params: 3,
        inputFormatter: [web3.extend.formatters.inputAddressFormatter, null, null]
      }]
    });

    web3.eth.getTokenTransfer("0x3dec3764c03AA5E7B570017C7Fd019109FA6e154", "0x86Fa049857E0209aa7D9e616F7eb3b3B78ECfdb0" 0, 1, (err, txs) => {
        console.log(txs);
    });

得到的是交易数组

    [{
      from: "0xd94c9ff168dc6aebf9b6cc86deff54f3fb0afc33",
      hash: "0xc6dbc9f6431830df8961b5f5305b66ba88b9084f0cf62e41eb7cc2791ad84591",
      timestamp: "1519267506",
      to: "0x3dec3764c03aa5e7b570017c7fd019109fa6e154",
      value: "178821000000000000000"
    }]

其中，`value`是最小单位，需要根据实际代币的`decimal`做处理才是实际交易金额

    web3.eth.getTokens("0x3dec3764c03AA5E7B570017C7Fd019109FA6e154", 0, 3, (err, tokens) => {
      console.log(tokens);
    })

得到的是一组地址，每个地址即为Token的合约地址

    [
      "0x6c22b815904165f3599f0a4a092d458966bd8024",
      "0x86fa049857e0209aa7d9e616f7eb3b3b78ecfdb0",
      "0xa0febbd88651ccca6180beefb88e3b4cf85da5be"
    ]
