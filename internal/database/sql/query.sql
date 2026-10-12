-- name: GetWallet :one
SELECT id,code,category,persona_version,account_version,state,created_at FROM wallets WHERE id=$1;
-- name: LockWallet :one
SELECT id,code,category,persona_version,account_version,state,created_at FROM wallets WHERE id=$1 FOR UPDATE;
-- name: ListWallets :many
SELECT id,code,category,persona_version,account_version,state,created_at FROM wallets ORDER BY created_at,code;
-- name: UpdateWallet :exec
UPDATE wallets SET state=$2, account_version=account_version+1 WHERE id=$1;
-- name: GetAction :one
SELECT body_hash,receipt FROM decision_actions WHERE wallet_id=$1 AND decision_id=$2 AND client_action_id=$3;
-- name: LatestSnapshot :one
SELECT body FROM market_snapshots WHERE asset_id=$1 ORDER BY captured_at DESC LIMIT 1;
