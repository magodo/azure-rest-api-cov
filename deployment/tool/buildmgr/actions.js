const {
    CosmosClient,
    OperationType
} = require("@azure/cosmos");

const fs = require('fs');
const cliProgress = require('cli-progress');

function buildDao(ep, key, dbName) {
    const db = new CosmosClient({
        endpoint: ep,
        key: key,
    }).database(dbName);

    return {
        build: db.container("build"),
        rp: db.container("rp"),
        path: db.container("path"),
    }
}

// chunks divide an array into sub-arraies where each one's length <= n.
function* chunks(arr, n) {
    for (let i = 0; i < arr.length; i += n) {
        yield arr.slice(i, i + n);
    }
}

/** 
 * The native bulk() method seems didn't hanlde the retry-after correctly.
 * safeBulk will permanently retry the operations that got 429 status code.
 * @param container - The container client
 * @param items - The []OperationInput of the bulk()
 */
async function safeBulk(container, items) {
    const bar = new cliProgress.SingleBar({}, cliProgress.Presets.shades_classic);
    bar.start(items.length, 0);
    for (let subitems of [...chunks(items, 100)]) {
        while (subitems.length > 0) {
            const resps = await container.items.bulk(subitems, {
                continueOnError: false,
            }, );
            let newSubItems = [];
            let maxRetryAfterMs = 100;
            resps.forEach((resp, index) => {
                if (resp.statusCode >= 200 && resp.statusCode < 300) {
                    return;
                }
                if (resp.statusCode != 429) {
                    throw Error(`bulk operation has failed item with response: ${resp}`);
                }
                if (resp.retryAfterMilliseconds > maxRetryAfterMs) {
                    maxRetryAfterMs = resp.retryAfterMilliseconds;
                }
                newSubItems.push(subitems[index]);
            });
            await new Promise(r => setTimeout(r, maxRetryAfterMs));
            bar.increment(subitems.length - newSubItems.length);
            subitems = newSubItems;
        }
    }
    bar.stop();
}

module.exports = {
    actionUpload: async (input, options, command) => {
        let opts = command.parent.opts();
        const dao = buildDao(opts.endpoint, opts.key, opts.database);

        const data = fs.readFileSync(input);
        const inputObj = JSON.parse(data);

        let uploadBuild = async () => {
            // Read from the container to get the latest build id, increment this by one and use it as the new build id.
            let buildId = '0';
            const {
                resources
            } = await dao.build.items.query(`SELECT c.id FROM c ORDER BY c._ts DESC OFFSET 0 LIMIT 1`).fetchAll();
            if (resources.length === 1) {
                buildId = (parseInt(resources[0].id) + 1).toString();
            }

            let buildObj = {
                "id": buildId,
                "comment": inputObj.comment,
                "ref": inputObj.ref,
            };

            await dao.build.items.create(buildObj)
            return buildId;
        };

        let uploadRp = async (buildId) => {
            let rpObjs = Object.keys(inputObj.resourceProviders).map(name => {
                return {
                    name,
                    buildId
                }
            });

            await safeBulk(dao.rp, rpObjs.map(obj => {
                return {
                    operationType: "Create",
                    resourceBody: obj,
                };
            }));

            // validate
            const {
                resources
            } = await dao.rp.items.readAll({
                partitionKey: buildId,
            }).fetchAll();

            if (resources.length !== rpObjs.length) {
                throw Error(`Validation failed, expected amount is ${rpObjs.length}, but got ${resources.length}`);
            }
        }

        let uploadPath = async (buildId) => {
            let pathObjs = [];
            for (let rp of Object.values(inputObj.resourceProviders)) {
                for (let path of Object.values(rp.paths)) {
                    pathObjs.push({
                        buildId: buildId,
                        rpName: rp.name,
                        name: path.name,
                        operations: path.operations,
                    });
                }
            }

            await safeBulk(dao.path, pathObjs.map(
                obj => {
                    return {
                        operationType: "Create",
                        resourceBody: obj,
                    };
                }
            ));

            // validate
            const {
                resources
            } = await dao.path.items.readAll({
                partitionKey: buildId,
            }).fetchAll();

            if (resources.length !== pathObjs.length) {
                throw Error(`Validation failed, expected amount is ${pathObjs.length}, but got ${resources.length}`);
            }
        }

        console.log("Uploading build info...");
        const buildId = await uploadBuild();
        console.log(`build id: ${buildId}`);

        console.log("Uploading RP info...");
        await uploadRp(buildId);

        console.log("Uploading API Path info...");
        await uploadPath(buildId);
    },

    actionDelete: async (buildId, options, command) => {
        let opts = command.parent.opts();
        const dao = buildDao(opts.endpoint, opts.key, opts.database);

        let deleteBuild = async (buildId) => {
            try {
                await dao.build.item(buildId, buildId).delete();
            } catch (err) {
                if (err.code !== 404) {
                    throw err;
                }
            }
        };

        let deleteRp = async (buildId) => {
            const {
                resources
            } = await dao.rp.items.readAll({
                partitionKey: buildId,
            }).fetchAll();

            await safeBulk(dao.rp, resources.map(
                item => {
                    return {
                        partitionKey: buildId,
                        operationType: "Delete",
                        id: item.id,
                    };
                }
            ));

            // validation
            {
                const {
                    resources
                } = await dao.rp.items.readAll({
                    partitionKey: buildId,
                }).fetchAll();
                if (resources.length !== 0) {
                    throw Error(`there are still ${resources.lenght} rp items exist`);
                }
            }
        }

        let deletePath = async (buildId) => {
            const {
                resources
            } = await dao.path.items.readAll({
                partitionKey: buildId,
            }).fetchAll();

            await safeBulk(dao.path, resources.map(
                item => {
                    return {
                        partitionKey: buildId,
                        operationType: "Delete",
                        id: item.id,
                    };
                }
            ));

            // validation
            {
                const {
                    resources
                } = await dao.path.items.readAll({
                    partitionKey: buildId,
                }).fetchAll();
                if (resources.length !== 0) {
                    throw Error(`there are still ${resources.length} path items exist`);
                }
            }
        }

        console.log("Deleting build info...");
        await deleteBuild(buildId);

        console.log("Deleting RP info...");
        await deleteRp(buildId);

        console.log("Deleting API Path info...");
        await deletePath(buildId);
    },
}