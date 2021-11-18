<p align="center"> <img src="https://user-images.githubusercontent.com/91356865/142333952-93557ec8-21be-40ca-a85a-7624fc731f7f.png" width="300"> 
</p> <p align="center"></p>

***

# gossip-propagation-d  
gossip-propagation-d は、同一ネットワークに接続された複数端末間でクラスターを形成します。クラスター内の端末間で、エッジ端末情報(デバイス名、IPアドレス、死活等)や、podの起動情報を取得・同期します。  
同期されたデータは、titaniadb-sentinelによって、titaniadb に書き込まれます。  
gossip-propagation-d は、コンテナ上で稼働せず、OSレイヤーで稼働します。  
gossip-propagation-d が OSレイヤーで稼働する理由は、エッジコンピューティング環境においてコンテナオーケストレーションシステムが単一障害点とならないようにするためです。  

![gossip-propagation-d](Documents/titaniadb_architecture2.PNG) 

## 依存関係
依存関係にあるマイクロサービスは、以下の通りです。 

- [distributed-service-discovery](https://github.com/latonaio/distributed-service-discovery)  
- [titaniadb-sentinel](https://github.com/latonaio/titaniadb-sentinel)  

## セットアップ方法
```
$ git clone git@github.com:latonaio/gossip-propagation-d.git -b v0.9.2 && cd gossip-propagation-d
$ make install
```

## 起動方法
### systemd 経由の起動
```
$ sudo systemctl start gossip-propagation-d.service
```

### マニュアルでの起動
```
$ gossip -j
```