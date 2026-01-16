# Telegram Trader

Use the golang telegram API and gswap/galachain typescript API (an example working bot using galachain gswap is here: https://github.com/cchuter/gswap-bot).

Trade assets on ston.fi and gswap.

In telegram, a user will be presented with these instructions after starting the bot:

```
🚀 Welcome to GalaSwap & STON.fi Trading Bot!

Trade tokens on GalaChain and TON blockchain with ease.

Available Commands:
/wallet - Connect your TonKeeper wallet
/balance - View your balances
/swap - Execute a token swap
/price - Check token prices
/alert - Set price alerts
/portfolio - View your portfolio
/orders - Manage automated orders
/help - Show all commands

Get started by connecting your wallet with /wallet
```

The user will be able to connect their TON wallet via Tonkeeper (app on iphone)

The user will be able to connect their GALA wallet via the gala wallet app (on iphone)

prices return valid token values retreived from either chain, TON or GALA (if the user doesn't specify it returns form both chains)

the TON token is GTON on galachain

the main pair we are interested in is:
on ston.fi: https://app.ston.fi/swap?chartVisible=false&ft=TON&tt=EQBadmOayy7_bD18skopfOZw2kmTgDdBhXPVsuTQq1lalaBV
on gswap: https://swap.gala.com/explore-balance/2954bd1e38c2dff11a2ee8798e2f59b99c206e64bef7bdf5c4ac9a91315e0d34

Give the ability to do spread arbs: Add add a "arb" command that looks at those proces and if they're .1% different does a trade of 50% of either asset (always eaving at least 10 gala or 1 TON). The goal is to buy when one token is less expensive on one chain and sell if its higher on the other - between the TON/GALA pair.
