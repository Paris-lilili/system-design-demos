## What this verifies
Kafka assigns a message to a partition by hashing its key (partition = hash(key) % numPartitions), so messages with the same key always land on the same partition. Kafka guarantees ordering **within** a partition, but **not across** partitions.

Verified two things:
1. Same key -> same partition -> events of one order stay strictly ordered (seq 1..5).
2. Cross-partition doesn't have global order: order_A was produced before order_B, but the consumer received order_B first — the cross-partition order does not follow production order.

Note: the default Balancer is RoundRobin (ignores the key), which scatters one order's events across partitions and breaks ordering. Setting `Balancer: &kafka.Hash{}` is what makes same-key-same-partition work.

## Verify and Output
1. Launch kafka container with topic, 3 partitions, 1 replication
```
❯ cd ~/Documents/projects/system-design-demos/payment-system/05-kafka-partition-ordering
** # docker compose via yaml file ** 
❯ docker compose up -d
[+] up 2/2
 ✔ Network 05-kafka-partition-ordering_default Created                                                                                                                        0.0s
 ✔ Container kafka-demo                        Started                                                                                                                        0.3s
❯ docker compose logs kafka | tail -30
kafka-demo  | ===> User
kafka-demo  | uid=1000(appuser) gid=1000(appuser) groups=1000(appuser)
kafka-demo  | ===> Configuring ...
kafka-demo  | Running in KRaft mode...
kafka-demo  | ===> Running preflight checks ... 
kafka-demo  | ===> Check if /var/lib/kafka/data is writable ...
kafka-demo  | ===> Running in KRaft mode, skipping Zookeeper health check...
kafka-demo  | ===> Using provided cluster id MkU3OEVBNTcwNTJENDM2Qk ...
kafka-demo  | ===> Launching ... 
kafka-demo  | ===> Launching kafka ... 
kafka-demo  | [2026-07-04 16:22:13,255] INFO Registered kafka:type=kafka.Log4jController MBean (kafka.utils.Log4jControllerRegistration$)
kafka-demo  | [2026-07-04 16:22:13,517] INFO Setting -D jdk.tls.rejectClientInitiatedRenegotiation=true to disable client-initiated TLS renegotiation (org.apache.zookeeper.common.X509Util)

❯ docker compose ps
NAME         IMAGE                         COMMAND                  SERVICE   CREATED          STATUS          PORTS
kafka-demo   confluentinc/cp-kafka:7.7.1   "/etc/confluent/dock…"   kafka     40 seconds ago   Up 39 seconds   0.0.0.0:9092->9092/tcp, [::]:9092->9092/tcp
❯ docker compose logs kafka | grep -i started | tail -5
kafka-demo  | [2026-07-04 16:22:14,517] INFO [ControllerServer id=1] Finished waiting for all of the SocketServer Acceptors to be started (kafka.server.ControllerServer)
kafka-demo  | [2026-07-04 16:22:15,061] INFO [BrokerServer id=1] Waiting for all of the SocketServer Acceptors to be started (kafka.server.BrokerServer)
kafka-demo  | [2026-07-04 16:22:15,061] INFO [BrokerServer id=1] Finished waiting for all of the SocketServer Acceptors to be started (kafka.server.BrokerServer)
kafka-demo  | [2026-07-04 16:22:15,062] INFO [BrokerServer id=1] Transition from STARTING to STARTED (kafka.server.BrokerServer)
kafka-demo  | [2026-07-04 16:22:15,063] INFO [KafkaRaftServer nodeId=1] Kafka Server started (kafka.server.KafkaRaftServer)

❯ go run main.go 

❯ docker exec kafka-demo kafka-topics --bootstrap-server localhost:9092 --describe --topic payment-events
Topic: payment-events   TopicId: GJlrHVWcSyeTjnfhS1Hfeg PartitionCount: 3       ReplicationFactor: 1    Configs: 
        Topic: payment-events   Partition: 0    Leader: 1       Replicas: 1     Isr: 1
        Topic: payment-events   Partition: 1    Leader: 1       Replicas: 1     Isr: 1
        Topic: payment-events   Partition: 2    Leader: 1       Replicas: 1     Isr: 1
```

2. send msg to kafka
2.1 Balancer default is RoundRobin, so event goes to 3 partitions one by one. the partition key becomes useless. But the goal is that all events of the same order should go to the same partition (to keep them ordered)
(ignore the timeout error, it is historical issue produced by kafka-console-consumer, when it logout, it print it)

```
❯ docker exec kafka-demo kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic payment-events \
  --from-beginning \
  --property print.partition=true \
  --property print.key=true \
  --timeout-ms 3000
Partition:2     order_A {"order_id":"order_A","seq":3,"type":"processing"}
Partition:2     order_B {"order_id":"order_B","seq":1,"type":"processing"}
Partition:2     order_B {"order_id":"order_B","seq":4,"type":"processing"}
Partition:1     order_A {"order_id":"order_A","seq":2,"type":"processing"}
Partition:1     order_A {"order_id":"order_A","seq":5,"type":"processing"}
Partition:1     order_B {"order_id":"order_B","seq":3,"type":"processing"}
Partition:0     order_A {"order_id":"order_A","seq":1,"type":"processing"}
Partition:0     order_A {"order_id":"order_A","seq":4,"type":"processing"}
Partition:0     order_B {"order_id":"order_B","seq":2,"type":"processing"}
Partition:0     order_B {"order_id":"order_B","seq":5,"type":"processing"}
[2026-07-06 16:01:53,884] ERROR Error processing message, terminating consumer process:  (kafka.tools.ConsoleConsumer$)
org.apache.kafka.common.errors.TimeoutException
Processed a total of 10 messages
```

2.2 [EXPECTED] After setting Balancer: &kafka.Hash{}, the partition is chosen by hashing the key (partition = hash(key) % numPartitions). Since the same key always hashes to the same value, messages with the same key always land on the same partition, and Kafka guarantees order within a partition.
Note: the console displays msg grouped by partition (print partition 1 then partition 0), so it clearly shows ordering within a partition, but it doesn't reveal the ccross-partition is disorder.
```
# clean all msg before re-sending msg
❯ docker exec kafka-demo kafka-topics --bootstrap-server localhost:9092 --delete --topic payment-events

❯ docker exec kafka-demo kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic payment-events \
  --from-beginning \
  --property print.partition=true \
  --property print.key=true \
  --timeout-ms 3000
Partition:1     order_B {"order_id":"order_B","seq":1,"type":"processing"}
Partition:1     order_B {"order_id":"order_B","seq":2,"type":"processing"}
Partition:1     order_B {"order_id":"order_B","seq":3,"type":"processing"}
Partition:1     order_B {"order_id":"order_B","seq":4,"type":"processing"}
Partition:1     order_B {"order_id":"order_B","seq":5,"type":"processing"}
Partition:0     order_A {"order_id":"order_A","seq":1,"type":"processing"}
Partition:0     order_A {"order_id":"order_A","seq":2,"type":"processing"}
Partition:0     order_A {"order_id":"order_A","seq":3,"type":"processing"}
Partition:0     order_A {"order_id":"order_A","seq":4,"type":"processing"}
Partition:0     order_A {"order_id":"order_A","seq":5,"type":"processing"}
[2026-07-06 19:45:18,633] ERROR Error processing message, terminating consumer process:  (kafka.tools.ConsoleConsumer$)
org.apache.kafka.common.errors.TimeoutException
Processed a total of 10 messages
```

3. Consuming msg from kafka

```
# clean all msg before re-consuming msg
❯ docker exec kafka-demo kafka-consumer-groups --bootstrap-server localhost:9092 --delete --group demo-consumer

❯ go run main.go
produced 10 msgs
topic= payment-events, partition= 1, key= order_B, value= {"order_id":"order_B","seq":1,"type":"processing"}
topic= payment-events, partition= 1, key= order_B, value= {"order_id":"order_B","seq":2,"type":"processing"}
topic= payment-events, partition= 1, key= order_B, value= {"order_id":"order_B","seq":3,"type":"processing"}
topic= payment-events, partition= 1, key= order_B, value= {"order_id":"order_B","seq":4,"type":"processing"}
topic= payment-events, partition= 1, key= order_B, value= {"order_id":"order_B","seq":5,"type":"processing"}
topic= payment-events, partition= 0, key= order_A, value= {"order_id":"order_A","seq":1,"type":"processing"}
topic= payment-events, partition= 0, key= order_A, value= {"order_id":"order_A","seq":2,"type":"processing"}
topic= payment-events, partition= 0, key= order_A, value= {"order_id":"order_A","seq":3,"type":"processing"}
topic= payment-events, partition= 0, key= order_A, value= {"order_id":"order_A","seq":4,"type":"processing"}
topic= payment-events, partition= 0, key= order_A, value= {"order_id":"order_A","seq":5,"type":"processing"}
```