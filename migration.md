# Finalizing Migration from `canine-provider`
This guide is only for storage provider that once used `canine-provider` in Jackal v3.

## Locating Storage
Locate your `storage` folder, by default it will be located at `$HOME/.jackal-storage/storage`. 

Whatever the directory above `storage` is, will be the absolute path to storage we use later. In the default case, it would be `$HOME/.jackal-storage`.

## Install Sequoia v1.1.0
```shell
git clone https://github.com/JackalLabs/sequoia.git
cd sequoia
git switch v1.1.0
make install
```

## Migrating
If you use any flags when using `sequoia start` please also apply these flags to the following command.
```shell
sequoia jprovd-salvage [ABSOLUTE-PATH-TO-STORAGE]
```

This command will bring all the `canine-provider` data into sequoia, and print out a list of file data to `salvage_record` in your `sequoia` root directory.

Please send this file to the Jackal Labs team (marston@jackallabs.io) or `@marstonc` on discord.

After this, you can continue running sequoia like normal.
