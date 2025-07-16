import { Gateway, Wallets } from 'fabric-network';
import FabricCAServices from 'fabric-ca-client';
import { join } from 'path';
import { buildCAClient, registerAndEnrollUser, enrollAdmin } from './CAUtil.js';
import { buildCCPOrg1, buildCCPOrg2, buildWallet } from './AppUtil.js';
import prompts from 'prompts';
import fs from 'fs';
import path from 'path';
import { Console } from 'console';

const channelName = 'mychannel';
const chaincodeName = 'chaincode';

const org1 = 'Org1MSP';
const org2 = 'Org2MSP';
const Org1UserId = 'app1User';
const Org2UserId = 'app2User';

const RED = '\x1b[31m\n';
const GREEN = '\x1b[32m\n';
const RESET = '\x1b[0m';

async function initGatewayForOrg1() {
	console.log(`${GREEN}--> Fabric client user & Gateway init: Using Org1 identity to Org1 Peer${RESET}`);
	// build an in memory object with the network configuration (also known as a connection profile)
	const ccpOrg1 = buildCCPOrg1();

	// build an instance of the fabric ca services client based on
	// the information in the network configuration
	const caOrg1Client = buildCAClient(FabricCAServices, ccpOrg1, 'ca.org1.example.com');

	// setup the wallet to cache the credentials of the application user, on the app server locally
	const walletPathOrg1 = join(process.cwd(), 'wallet', 'org1');
	const walletOrg1 = await buildWallet(Wallets, walletPathOrg1);

	// in a real application this would be done on an administrative flow, and only once
	// stores admin identity in local wallet, if needed
	await enrollAdmin(caOrg1Client, walletOrg1, org1);
	// register & enroll application user with CA, which is used as client identify to make chaincode calls
	// and stores app user identity in local wallet
	// In a real application this would be done only when a new user was required to be added
	// and would be part of an administrative flow
	await registerAndEnrollUser(caOrg1Client, walletOrg1, org1, Org1UserId, 'org1.department1');

	try {
		// Create a new gateway for connecting to Org's peer node.
		const gatewayOrg1 = new Gateway();
		//connect using Discovery enabled
		await gatewayOrg1.connect(ccpOrg1,
			{ wallet: walletOrg1, identity: Org1UserId, discovery: { enabled: true, asLocalhost: true } });

		return gatewayOrg1;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org1: ${error}`);
		process.exit(1);
	}
}

async function initGatewayForOrg2() {
	console.log(`${GREEN}--> Fabric client user & Gateway init: Using Org2 identity to Org2 Peer${RESET}`);
	const ccpOrg2 = buildCCPOrg2();
	const caOrg2Client = buildCAClient(FabricCAServices, ccpOrg2, 'ca.org2.example.com');

	const walletPathOrg2 = join(process.cwd(), 'wallet', 'org2');
	const walletOrg2 = await buildWallet(Wallets, walletPathOrg2);

	await enrollAdmin(caOrg2Client, walletOrg2, org2);
	await registerAndEnrollUser(caOrg2Client, walletOrg2, org2, Org2UserId, 'org2.department1');

	try {
		// Create a new gateway for connecting to Org's peer node.
		const gatewayOrg2 = new Gateway();
		await gatewayOrg2.connect(ccpOrg2,
			{ wallet: walletOrg2, identity: Org2UserId, discovery: { enabled: true, asLocalhost: true } });

		return gatewayOrg2;
	} catch (error) {
		console.error(`Error in connecting to gateway for Org2: ${error}`);
		process.exit(1);
	}
}

async function AddBalance(contract) {
    const questionsForAddMoneyBalance = [{
        type: 'number',
        name: 'amount',
        message: 'Amount of money to add: '
    }]
    
    await prompts(questionsForAddMoneyBalance)
        .then(async response => {
            const transientDataJSON = JSON.stringify(response);
            const transientDataBuffer = Buffer.from(transientDataJSON);

            const transientData = {
                amount: transientDataBuffer
            }

            try {
                await contract
                    .createTransaction('AddBalance')
                    .setTransient(transientData)
                    .submit()

                console.log('[-] Done.')
            }
            catch(error) {
                console.log("[-] Error: ", error.message)
            }
        })
        .catch(error => {
            console.log("[-] Error: ", error)
        })
}

async function AddItem(contract, org) {
    const questionsForAddItem = [{
        type: 'text',
        name: 'name',
        message: 'Name of the item: '
    }, {
        type: 'number',
        name: 'count',
        message: 'Count of the item: ',
        initial: 0
    }, {
        type: 'number',
        name: 'price',
        message: 'Price of the item: ',
    }]

    await prompts(questionsForAddItem)
        .then(async response => {
            const transientDataJSON = JSON.stringify(response);
            const transientDataBuffer = Buffer.from(transientDataJSON);

            const transientData = {
                item: transientDataBuffer
            }

            var tx = await contract.createTransaction('AddItem')
            if(org == 'org1') {
                tx.setEndorsingOrganizations(org1)
            } else if(org == 'org2') {
                tx.setEndorsingOrganizations(org2)
            }

            tx.setTransient(transientData)

            try {
                await tx.submit()
                console.log('[-] Done.')
            }
            catch(error) {
                console.log("[-] Error: ", error.message)
            }
        })
        .catch(error => {
            console.log("[-] Error: ", error)
        })
}

async function AddItemToMarket(contract, org) {
    const questionsForAddItemToMarket = [{
        type: 'text',
        name: 'name',
        message: 'Name of the item: '
    }, {
        type: 'number',
        name: 'price',
        message: 'Price of the item: ',
    }]

    await prompts(questionsForAddItemToMarket)
        .then(async response => {
            var tx = await contract.createTransaction('AddToMarket')
            if(org == 'org1') {
                tx.setEndorsingOrganizations(org1)
            } else if(org == 'org2') {
                tx.setEndorsingOrganizations(org2)
            }

            try {
                await tx.submit(response["name"] , response["price"])
                console.log('[-] Done.')
            }
            catch(error) {
                console.log("[-] Error: ", error.message)
            }
        })
        .catch(error => {
            console.log("[-] Error: ", error)
        })
}

async function QueryBalance(contract) {
    try {
        var result = await contract.evaluateTransaction('GetBalance')
        console.log('[-] Output: ', result.toString())
    }
    catch(error) {
        console.log("[-] Error: ", error.message)
    }
}

async function GetItem(contract) {
    try {
        var result = await contract.evaluateTransaction('GetItem')
        if(result.toString() == '') {
            console.log('[-] No items in inventory.')
        }
        else {
            console.table(JSON.parse(result.toString()))
        }
    }
    catch(error) {
        console.log("[-] Error: ", error.message)
    }
}

async function GetItemsInMarket(contract) {
    try {
        var result = await contract.evaluateTransaction('GetItemsInMarket')
        if(result.toString() == '') {
            console.log('[-] No items in inventory.')
        }
        else {
            console.table(JSON.parse(result.toString()))
        }
    }
    catch(error) {
        console.log("[-] Error: ", error.message)
    }
}

async function AddToWishlist(wishlist, contract, org) {
    const questionsForBuyFromMarket = [{
        type: 'text',
        name: 'name',
        message: 'Name of the item: '
    }]

    await prompts(questionsForBuyFromMarket)
        .then(async response => {
            try {
                var tx = await contract.createTransaction('BuyFromMarket')
               
                if(org == 'org1') {
                    await tx.submit(org2 + "_" + response["name"])
                } else if(org == 'org2') {
                    await tx.submit(org1 + "_" + response["name"])
                }
                
                console.log('[-] Item found in marketplace and bought.')
            }
            catch(error) {
                console.log("[-] Cannot buy the item now because of the error: ", error.message)
                console.log("[-] Adding the item to wishlist.")
                wishlist.push(response["name"])
            }
        })
        .catch(error => {
            console.log("[-] Error: ", error)
        })
}

function readWishlistHistory(org) {
    const dir = path.join(process.cwd(), '/wallet/', org);

    if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir);
    }

    const filePath = path.join(dir, 'wishlist.txt');
    if (!fs.existsSync(filePath)) {
        fs.writeFileSync(filePath, '');
        console.log('[-] Created wishlist history file at ', filePath)
    } else {
        console.log('[-] Wishlist history file found at ', filePath)
    }

    const fileData = fs.readFileSync(filePath, 'utf8').split('\n');
    fileData.pop()
    console.log('[-] Wishlist History: ', fileData)
    return fileData
}

function SaveWishlistHistory(wishlist, org) {
    const dir = path.join(process.cwd(), '/wallet/', org);
    const filePath = path.join(dir, 'wishlist.txt');
    var fileData = ''
    wishlist.forEach(item => {
        fileData += item + '\n'
    })
    fs.writeFileSync(filePath, fileData);
    console.log('[-] Wishlist saved at ', filePath)
}

async function main() {
    const terminalPrompt = [{
        type: 'text',
        name: 'command',
        message: '$> ',
        initial: 'HELP'
    }, ]

    var contract
    if(process.argv[2] == 'org1') {
        /** ******* Fabric client init: Using Org1 identity to Org1 Peer ******* */
        const gatewayOrg1 = await initGatewayForOrg1();
        const networkOrg1 = await gatewayOrg1.getNetwork(channelName);
        const contractOrg1 = networkOrg1.getContract(chaincodeName);
        
        contract = contractOrg1
    } else if(process.argv[2] == 'org2') {
        /** ******* Fabric client init: Using Org2 identity to Org2 Peer ******* */
        const gatewayOrg2 = await initGatewayForOrg2();
        const networkOrg2 = await gatewayOrg2.getNetwork(channelName);
        const contractOrg2 = networkOrg2.getContract(chaincodeName);

        contract = contractOrg2
    } else {
        console.log('[-] Please specify org1 or org2')
        process.exit()
    }

    await contract.createTransaction('InitLedger').submit();

    console.log('\n')
    var wishlist = readWishlistHistory(process.argv[2])

    let handlingAnEvent = false
    const listener = async (event) => {
        // handlingAnEvent = true
        console.log('\n')
        if(event.eventName == 'item_added') {
            const asset = JSON.parse(event.payload.toString());
            if(wishlist.includes(asset["Name"])) {
                console.log('[+] An item present in the wishlist was added in the marketplace.')
                console.table(asset)
                console.log('[+] Proceeding to buy the item')

                try {
                    var tx = await contract.createTransaction('BuyFromMarket')

                    await tx.submit(asset["ID"])
                    console.log('\n[+] Item bought. Press "enter" key...')
                    process.stdin.write('\n')

                    wishlist = wishlist.filter(item => item != asset["Name"])
                }
                catch(error) {
                    console.log("[-] Error: ", error.message)
                }
            }
        }
        else {
            console.log(`${RED} [+] Unknown event detected with payload:${RESET}`)
            console.log(event.payload.toString())
        }
        // handlingAnEvent = false
    }

    try {
        contract.addContractListener(listener)
    } catch(error) {
        console.log("[-] Error in adding event listener: ", error.message)
    }

    /** ****** terminal greeter ****** **/
    console.log(`${GREEN} ^^^^ APPLICATION TERMINAL SETUP COMPLETED ^^^^${RESET}`);
	while (1) {
        while(handlingAnEvent) {}
        await prompts(terminalPrompt)
            .then(async response => {
                const actualCommand = response.command.trim().toUpperCase()

                switch(actualCommand) {
                    case 'ADD_MONEY': 
                        await AddBalance(contract)
                        break
                    case 'ADD_ITEM': 
                        await AddItem(contract, process.argv[2])
                        break
                    case 'QUERY_BALANCE': 
                        await QueryBalance(contract)
                        break
                    case 'GET_ITEM': 
                        await GetItem(contract)
                        break
                    case 'ENLIST_ITEM':
                        await AddItemToMarket(contract, process.argv[2])
                        break
                    case 'ALL_ITEMS':
                        await GetItemsInMarket(contract)
                        break
                    case 'WISHLIST':
                        await AddToWishlist(wishlist, contract, process.argv[2])
                        break
                    case 'EXIT':
                        console.log('[-] Saving wishlist and exiting gracefully.')
                        SaveWishlistHistory(wishlist, process.argv[2])
                        process.exit()
                    default:
                        console.table([
                            {
                                command: 'ADD_MONEY',
                                description: 'adds money to balance',
                            },
                            {
                                command: 'ADD_ITEM',
                                description: 'adds an item to inventory',
                            },
                            {
                                command: 'QUERY_BALANCE',
                                description: 'retrieves current balance',
                            },
                            {
                                command: 'GET_ITEM',
                                description: 'retreives details about an item',
                            },
                            {
                                command: 'ENLIST_ITEM',
                                description: 'add item to marketplace',
                            },
                            {
                                command: 'ALL_ITEMS',
                                description: 'get all items in marketplace',
                            },
                            {
                                command: 'WISHLIST',
                                description: 'add item to wishlist and possibly buy them later',
                            },
                            {
                                command: 'EXIT',
                                description: 'exit the program',
                            },
                        ])  
                }
            })
            .catch(error => {
                console.log("error: ", error)
            })
    }
}

const errorLogFile = 'logs.txt'
const errorStream = fs.createWriteStream(errorLogFile)
process.stderr.write = errorStream.write.bind(errorStream)

main().catch((error) => {
    console.log(`[-] Error in application: ${error}`);
});