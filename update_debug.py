filepath = r'd:\project\golang\kjt_go_dana\internal\dana\sdk_client.go'
with open(filepath, 'r', encoding='utf-8') as f:
    content = f.read()

old = '\t// TransactionDate is mandatory and MUST be in GMT+7 (+07:00) format for DANA\n\tloc := time.FixedZone("WIB", 7*3600)\n\tsdkReq.SetTransactionDate(time.Now().In(loc).Format("2006-01-02T15:04:05+07:00"))\n\n\t// Call SDK'

new = '\t// TransactionDate is mandatory and MUST be in GMT+7 (+07:00) format for DANA\n\tloc := time.FixedZone("WIB", 7*3600)\n\tsdkReq.SetTransactionDate(time.Now().In(loc).Format("2006-01-02T15:04:05+07:00"))\n\n\t// ===== DEBUG: Dump full request payload =====\n\treqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")\n\tlog.Printf("[DANA SDK] ===== TRANSFER TO DANA REQUEST PAYLOAD =====")\n\tlog.Printf("[DANA SDK] %s", string(reqJSON))\n\tlog.Printf("[DANA SDK] ============================================")\n\n\t// Call SDK'

content = content.replace(old, new)

with open(filepath, 'w', encoding='utf-8') as f:
    f.write(content)

print('sdk_client.go updated with debug logging')
