## Go Ethereum

在[go-ethereum](https://github.com/ethereum/go-ethereum)的基础上，增加getAddressTransactions用于获取任意地址的交易记录。

在`erc20`分支中，增加了getTokenTransfer和getTokens两个函数用于获取erc20代币的交易记录

在`geth console`中

    eth.getAddressTransactions(address, start, end)

在`web3.js`中

    const Web3 = require("web3")
    const net = require("net")
    const web3 = new Web3("/home/user/.ethereum/geth.ipc", net);

    web3.eth.extend({
        methods: [{
            name: "getAddressTransactions",
            call: 'eth_getAddressTransactions',
            params: 3,
            inputFormatter: [web3.extend.formatters.inputAddressFormatter, null, null]
        }]
    });

    web3.eth.getAddressTransactions("0xA1E4380A3B1f749673E270229993eE55F35663b4", 0, 1, (err, txs) => {
        console.log(txs);
    });

得到的是交易数组

    [ 
      { 
        from: '0xa1e4380a3b1f749673e270229993ee55f35663b4',
        to: '0xe28e72fcf78647adce1f1252f240bbfaebd63bcc',
        value: 800000000000000000000,
        gasPrice: 57935965411,
        gasUsed: 114110,
        blockHash: '0x2ee3b5b1eab71c30cbf1e84f2d25b87e5cdd9a69f530a588f7081702baa19826',
        blockNumber: 116219,
        status: 0,
        timestamp: 1440070385,
        hash: '0xf0762a50f9f591e61f65d86787e3b8e4c0703badb5d571707d840da575c0cf57',
        kind: 0 
      },
      { 
        from: '0xa1e4380a3b1f749673e270229993ee55f35663b4',
        to: '0xcde4de4d3baa9f2cb0253de1b86271152fbf7864',
        value: 0,
        gasPrice: 50000000000000,
        gasUsed: 180198,
        blockHash: '0x3f1afe224d29254e675adeccef2358d53887427037ba1421dc6efe785ce4eb40',
        blockNumber: 49392,
        status: 0,
        timestamp: 1438972859,
        hash: '0xef1b643d0cfb01321dfda1115fe8fb3181d9b592a79dfb75d1281117180d2d65',
        kind: 1 
      }
    ]

其中，`status`等于1时，代表合约执行成功，为0时代表失败

`kind`为0时，表示这笔交易为以太币转账或者调用智能合约（可通过`to`是否为合约地址区分），`kind`为1时，表示这笔交易为`Contract Creation`,`to`即为创建的合约地址

`hash`为此交易的哈希，`blockHash`为交易所在区块的哈希，`timestamp`也是区块的时间戳
