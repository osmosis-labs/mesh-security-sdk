async function getFaucet(address) {
    const body = {
      address: address,
      amount: '100',
    }

    let options = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json;charset=utf-8',
      },
      body: JSON.stringify(body),
    }

    // See https://docs.bandchain.org/develop/api-endpoints#laozi-testnet-5
    let response = await fetch(`https://laozi-testnet6.bandchain.org/faucet`, options)

    console.log(response)
  }

const address = process.argv[2];
getFaucet(address)
