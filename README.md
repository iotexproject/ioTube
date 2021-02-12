[![tests](https://github.com/iotexproject/ioTube/workflows/tests/badge.svg)](https://github.com/iotexproject/ioTube/actions?query=workflow%3Atests)

<p align="center">
  <img src="https://github.com/iotexproject/ioTube/blob/master/ioTube.png" width="480px">
</p>

<p>
  <strong>A multi-assets, fully decentralized and bidirectional bridge for exchanging ERC20/XRC20 tokens between Ethereum and IoTeX.</strong>
  
ioTube is a decentralized cross-chain bridge that enables the bidirectional exchange of digital assets (e.g., tokens, stable coins) between IoTeX and other blockchain networks. The <a href="https://member.iotex.io/tools/iotex">first version</a> of ioTube was built by core-dev around 2019 June to facilitate the swap of IOTX-ERC20 and IOTX until now, and this released version generalizes the original ioTube to support multiple assets on Ethereum & IoTeX blockchains. In the future, we plan to add support more blockchain networks to increase the reach and impact of ioTube.
</p>

<h3>
      <a href="https://github.com/iotexproject/ioTube#deployement">Deployement</a>
      <span> | </span>
      <a href="https://github.com/iotexproject/ioTube#usage">Usage</a>
      <span> | </span>
      <a href="https://github.com/iotexproject/ioTube/tree/master/docs">Documentation</a>
</h3>

&nbsp;

## Deployment
Different from traditional bridges, ioTube comes with two components:
- **a golang service** witnessing what has happened on both changes and relay the finalized information
- **a set of smart contracts** pre-deployed on both chains letting the legitimate witnesses relay information back and forth to facilitate cross-chain transferring of assets

### Deploy Contracts on IoTeX/Ethereum
* Deploy a TokenSafe `ts`
* Deploy a TokenList `tl1`, which stores tokens using `ts`
* Deploy a MinterPool `mp`
* Deploy a TokenList `tl2`, which stores tokens using `mp`
* Deploy a WitnessList `wl`
* Deploy a TokenCashier `tc` with `ts`, `tl1`, and `tl2`
* Deploy a TransferValidator `tv` with `ts`, `mp`, `tl1`, `tl2`, and `wl`
* Transfer ownership of `mp` to `tv`
* Transfer ownership of `ts` to `tv`
* Add witnesses to `wl`
* Add tokens to `tl1` or `tl2`

### Join as a Relayer

1. Edit `witness-service/relayer-config-iotex.yaml` and `witness-service/relayer-config-ethereum.yaml` to fill in the following fields:
* privateKey
* clientURL
* validatorContractAddress

2. start containers
```
cd witness-service
./start_relayer.sh
```

### Join as a Witness

1. Edit `witness-service/witness-config-iotex.yaml` and `witness-service/witness-config-ethereum.yaml` to fill in the following fields:
* privateKey
* clientURL
* validatorContractAddress
* cashierContractAddress

2. start containers
```
cd witness-service
./start_witness.sh
```

3. Clean up everything by running
```
./clean-all.sh
```

### Transfer assets between IoTeX and Ethereum

Please use dApp ioTube https://tube.iotex.io. Please note that the service is still in beta mode.  

#### From ERC20 to XRC20

1. open https://tube.iotex.io/eth in a metamask installed browser. (eg. Firefox/Chrome/Brave + metamask) 
2. Choose supported ERC20 token from the list.
3. Enter the amount. 
4. Click `Approve` button to approve ERC20 token transfer and sign on metamask.
5. Click `Convert` button and sign on metamask.
6. After `12` confirmations of Ethereum network and `2/3 + 1` confirmations from witnesses, the XRC20 tokens will be minted and sent to your IoTeX address.
7. You can add the token to your <a href="http://iopay.iotex.io/">ioPay</a> to see and use them.

#### From XRC20 to ERC20

1. open https://tube.iotex.io/iotx in ioPay desktop supported broswers (eg. Chrome/Firefox/Brave with ioPay desktop installed) or ioPay Android/iOS.
2. Choose supported XRC20 token from the list.
3. Enter the amount.
4. Click `Approve` button to approve XRC20 token transfer and sign on ioPay.
5. Click `Convert` button and sign on ioPay.
6. After `1` confirmation of IoTeX network and `2/3 + 1` confirmations from witnesses, the XRC20 token will be burnt and ERC20 token will be sent to your ETH wallet.

#### Fees
Tube fees: `0`

Network fees: 

1. from ERC20 to XRC20: `0`
2. from XRC20 to ERC20: `2000 IOTX` (to cover the high ETH gas fee for witnesses)


### Add an ERC20 token to ioTube
* Add this token to `tl` on Ethereum
* Deploy a ShadowToken and add it to `tl` on IoTeX


## Security
- Each witness needs to be registered to `WitnessList` contract on both chains based on PoA (Proof-of-Authority).
- Each transferring of assets from one chain to another needs the endorsement from more than 2/3 of the registered witnesses; otherwise, it will not be successful.
- IoTeX has instant finality, meaning, to a witness, one confirmed block indicates finalization of a `burn` event, while on the Ethereum side, a witness needs to wai 12 blocks before making any endorsement.
- To launch ioTube reliably, we have limited the min/max of the asset that can be moved around. These limits can be lifted once ioTube gets more stress validated.

## Gas Costs
Gas fees on IoTeX are negligible, both for bridge maintenance and for asset transfer. The estimated gas fees on Ethereum side are:
- To transfer token from Ethereum to IoTeX: ~43,952 gas to set allowance and ~151,122 to lock it;
- To transfer token back from IoTeX to ETH: ~422,525 gas to unlock the token.

## Contract addresses

Contracts on IoTeX 

- Wrapped IOTX: io15qr5fzpxsnp7garl4m7k355rafzqn8grrm0grz
- Token Safe: io1cj3f498390srqv765nnvaxuk0rpxyadzpfjz75
- Minter Pool: io1g7va274ltufv5nh4xawfmt0clel6tfz58p7n5r
- Standard Token List: io1t89whrwyfr0supctsqcx9n7ex5dd8yusfqhyfz 
- Proxy Token List: io1dn8nqk3pmmll990xz6a94fpradtrljxmmx5p8j 
- Witness List: io1hezp6d7y3246c5gklnnkh0z95qfld4zdsphhsw
- Token Cashier: io1gsr52ahqzklaf7flqar8r0269f2utkw9349qg8
- Transfer Validator: io1dwaxh2ml4fd2wg8cg35vhfsgdcyzrczffp3vus

Contacts on Ethereum

- WETH: 0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2
- Standard Token List: 0x7c0bef36e1b1cbeb1f1a5541300786a7b608aede
- Proxy Token List: 0x73ffdfc98983ad59fb441fc5fe855c1589e35b3e
- Witness List: 0x8598dF1Ec0ac7dfBa802f4bDD93A6B93bd0AD83f 
- Token Safe: 0xc2e0f31d739cb3153ba5760a203b3bd7c27f0d7a 
- Minter Pool: 0x964f4f19bc823e72cc1f806021937cfc06f63b45 
- Token Cashier: 0xa0fd7430852361931b23a31f84374ba3314e1682
- Transfer Validator: 0xd8165188ccc135b3a3b2a5d2bc3af9d94753d955

## Current Supported Tokens

- WETH (0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2) - min=0.1 max=1000 

ioETH contract (io1qfvgvmk6lpxkpqwlzanqx4atyzs86ryqjnfuad)

- BUSD (0x4Fabb145d64652a948d72533023f6E7A623C7C53) - min=1 max=400000

ioBUSD contract (io14nhfkywdfvl40evgsqnh43ev33q6he8yez8c8a)

- UNISWAP (0x1f9840a85d5af5bf1d1762f925bdaddc4201f984) - min=0.1 max=80000

ioUNI contract (io1ahh04jn2zkqlug6feh7rpq75a75p3rj42hp4ch)

- PAXG (0x45804880de22913dafe09f4980848ece6ecbaf78) - min=0.01 max=200

ioPAXG contract (io19fsq8e9krrlng4ay5gyq6q5tqfym28yq9ly0fz)

