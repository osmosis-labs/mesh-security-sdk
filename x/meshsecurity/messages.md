# Active to tombstone
```json
{
  "valset_update": {
    "additions": [],
    "removals": [
      "cosmosvaloper1qpr9kv5wyatfuefj3f9xklm87es8yayd2qls7d"
    ],
    "updated": [
      {
        "address": "cosmosvaloper1qpr9kv5wyatfuefj3f9xklm87es8yayd2qls7d",
        "commission": "1.000000000000000000",
        "max_commission": "1.000000000000000000",
        "max_change_rate": "1.000000000000000000"
      }
    ],
    "jailed": [
      "cosmosvaloper1qpr9kv5wyatfuefj3f9xklm87es8yayd2qls7d"
    ],
    "unjailed": [],
    "tombstoned": [
      "cosmosvaloper1qpr9kv5wyatfuefj3f9xklm87es8yayd2qls7d"
    ]
  }
}
```

# Active to jailed
```json
{
  "valset_update": {
    "additions": [],
    "removals": [
      "cosmosvaloper1cdwwskxm5qjsgetkxlv95p5jshz05gn320z30t"
    ],
    "updated": [
      {
        "address": "cosmosvaloper1cdwwskxm5qjsgetkxlv95p5jshz05gn320z30t",
        "commission": "0.000000000000000000",
        "max_commission": "1.000000000000000000",
        "max_change_rate": "1.000000000000000000"
      }
    ],
    "jailed": [
      "cosmosvaloper1cdwwskxm5qjsgetkxlv95p5jshz05gn320z30t"
    ],
    "unjailed": [],
    "tombstoned": []
  }
}

```
# Jailed to active
```json
{
  "valset_update": {
    "additions": [
      {
        "address": "cosmosvaloper13kdr4felug9grnswvlrtegxrlvh8ks724sfugn",
        "commission": "0.000000000000000000",
        "max_commission": "1.000000000000000000",
        "max_change_rate": "1.000000000000000000"
      }
    ],
    "removals": [],
    "updated": [],
    "jailed": [],
    "unjailed": [
      "cosmosvaloper13kdr4felug9grnswvlrtegxrlvh8ks724sfugn"
    ],
    "tombstoned": []
  }
}
```

## Jailed to removed
3 differet events, unlikely at the same height
* jailed:
```json
{
  "valset_update": {
    "additions": [],
    "removals": [
      "cosmosvaloper18l39958dn3ntwdyz87ex3xa5l3mp02ym5r03xe"
    ],
    "updated": [
      {
        "address": "cosmosvaloper18l39958dn3ntwdyz87ex3xa5l3mp02ym5r03xe",
        "commission": "0.000000000000000000",
        "max_commission": "1.000000000000000000",
        "max_change_rate": "1.000000000000000000"
      }
    ],
    "jailed": [
      "cosmosvaloper18l39958dn3ntwdyz87ex3xa5l3mp02ym5r03xe"
    ],
    "unjailed": [],
    "tombstoned": []
  }
}
```
* new validator gets slot
```json
{
  "valset_update": {
    "additions": [
      {
        "address": "cosmosvaloper17w36un4m6s848secp045zlwztreeh40z0s6ffu",
        "commission": "0.000000000000000000",
        "max_commission": "1.000000000000000000",
        "max_change_rate": "1.000000000000000000"
      }
    ],
    "removals": [],
    "updated": [],
    "jailed": [],
    "unjailed": [],
    "tombstoned": []
  }
}

```
* unjailed but not back in active set
```json
{
  "valset_update": {
    "additions": [],
    "removals": [],
    "updated": [],
    "jailed": [],
    "unjailed": [
      "cosmosvaloper18l39958dn3ntwdyz87ex3xa5l3mp02ym5r03xe"
    ],
    "tombstoned": []
  }
}
```

