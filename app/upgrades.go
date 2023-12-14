package app

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	//runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
)

// RegisterUpgradeHandlers registers upgrade handlers.

func (app App) RegisterUpgradeHandlers() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	app.StickyFingers(upgradeInfo)
}
func (app *App) StickyFingers(_ upgradetypes.Plan) {
	planName := "StickyFingers"
	app.UpgradeKeeper.SetUpgradeHandler(
		planName,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("Cosmos-SDK v0.50.x is here...")
			// Print the modules with their respective ver.
			for moduleName, version := range fromVM {
				app.Logger().Info(fmt.Sprintf("Module: %s, Version: %d", moduleName, version))

			}
			// DEVNET-6 FIX: new consensus params keeper using the wrong key again and move the data into the consensus params keeper with the right key
			storesvc := runtime.NewKVStoreService(app.GetKey("upgrade"))
			consensuskeeper := consensuskeeper.NewKeeper(
				app.appCodec,
				storesvc,
				app.AccountKeeper.GetAuthority(),
				runtime.EventService{},
			)

			params, err := consensuskeeper.ParamsStore.Get(ctx)
			app.Logger().Info("Getting the params into the Consensus params keeper...")
			if err != nil {
				return nil, err
			}

			err = app.ConsensusParamsKeeper.ParamsStore.Set(ctx, params)
			app.Logger().Info("Setting the params into the Consensus params keeper...")
			if err != nil {
				return nil, err
			}

			return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		},
	)
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == planName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			Added: []string{
				circuittypes.ModuleName,
				"feeibc",
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

}
