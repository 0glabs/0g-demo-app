async function main() {
  const storeAbi = await ethers.getContractFactory("Store");
  const store = await storeAbi.deploy();
  console.log("store deployed at:", store.address);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
