const {
    actionUpload,
    actionDelete
} = require('./actions');

var pjson = require('./package.json');

async function main() {
    const {
        program
    } = require('commander');
    program.version(pjson.version);
    program
        .requiredOption('-e, --endpoint <endpoint>', 'the cosmosdb endpoint')
        .requiredOption('-k, --key <key>', 'the cosmosdb account master key')
        .requiredOption('-d, --database <database>', 'the name of the cosmosdb database')

    program
        .command('upload <input>')
        .description('Upload the build result from `azure-rest-api-cov feeddb` to cosmosdb')
        .action(actionUpload);
    program
        .command('delete <buildId>')
        .description('Delete the items for a certain build from cosmosdb')
        .action(actionDelete);

    await program.parseAsync(process.argv);
}

main().catch((err) => {
    console.log(err);
    process.exit(1);
});