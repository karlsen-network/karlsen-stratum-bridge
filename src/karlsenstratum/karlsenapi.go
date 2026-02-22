package karlsenstratum

import (
	"context"
	"fmt"
	"time"

	"github.com/karlsen-network/karlsen-stratum-bridge/v2/src/gostratum"
	"github.com/karlsen-network/karlsend/v2/app/appmessage"
	"github.com/karlsen-network/karlsend/v2/infrastructure/network/rpcclient"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type KarlsenApi struct {
	address       string
	blockWaitTime time.Duration
	logger        *zap.SugaredLogger
	karlsend      *rpcclient.RPCClient
	connected     bool
}

func NewKarlsenAPI(address string, blockWaitTime time.Duration, logger *zap.SugaredLogger) (*KarlsenApi, error) {
	client, err := rpcclient.NewRPCClient(address)
	if err != nil {
		return nil, err
	}

	return &KarlsenApi{
		address:       address,
		blockWaitTime: blockWaitTime,
		logger:        logger.With(zap.String("component", "karlsenapi:"+address)),
		karlsend:      client,
		connected:     true,
	}, nil
}

func (ks *KarlsenApi) Start(ctx context.Context, blockCb func()) {
	ks.waitForSync(true)
	go ks.startBlockTemplateListener(ctx, blockCb)
	go ks.startStatsThread(ctx)
}

func (ks *KarlsenApi) startStatsThread(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			ks.logger.Warn("context cancelled, stopping stats thread")
			return
		case <-ticker.C:
			dagResponse, err := ks.karlsend.GetBlockDAGInfo()
			if err != nil {
				ks.logger.Warn("failed to get network hashrate from karlsen, prom stats will be out of date", zap.Error(err))
				continue
			}
			response, err := ks.karlsend.EstimateNetworkHashesPerSecond(dagResponse.TipHashes[0], 1000)
			if err != nil {
				ks.logger.Warn("failed to get network hashrate from karlsen, prom stats will be out of date", zap.Error(err))
				continue
			}
			RecordNetworkStats(response.NetworkHashesPerSecond, dagResponse.BlockCount, dagResponse.Difficulty)
		}
	}
}

func (ks *KarlsenApi) reconnect() error {
	if ks.karlsend != nil {
		return ks.karlsend.Reconnect()
	}

	client, err := rpcclient.NewRPCClient(ks.address)
	if err != nil {
		return err
	}
	ks.karlsend = client
	return nil
}

func (s *KarlsenApi) waitForSync(verbose bool) error {
	if verbose {
		s.logger.Info("checking karlsend sync state")
	}
	for {
		clientInfo, err := s.karlsend.GetInfo()
		if err != nil {
			return errors.Wrapf(err, "error fetching server info from karlsend @ %s", s.address)
		}
		if clientInfo.IsSynced {
			break
		}
		s.logger.Warn("Karlsen is not synced, waiting for sync before starting bridge")
		time.Sleep(5 * time.Second)
	}
	if verbose {
		s.logger.Info("karlsend synced, starting server")
	}
	return nil
}

func (s *KarlsenApi) startBlockTemplateListener(ctx context.Context, blockReadyCb func()) {
	var blockReadyChan chan bool
	restartChannel := true
	ticker := time.NewTicker(s.blockWaitTime)
	for {
		if err := s.waitForSync(false); err != nil {
			s.logger.Error("error checking karlsend sync state, attempting reconnect: ", err)
			if err := s.reconnect(); err != nil {
				s.logger.Error("error reconnecting to karlsend, waiting before retry: ", err)
				time.Sleep(5 * time.Second)
			}
			restartChannel = true
		}
		if restartChannel {
			blockReadyChan = make(chan bool)
			err := s.karlsend.RegisterForNewBlockTemplateNotifications(func(_ *appmessage.NewBlockTemplateNotificationMessage) {
				blockReadyChan <- true
			})
			if err != nil {
				s.logger.Error("fatal: failed to register for block notifications from karlsen")
			} else {
				restartChannel = false
			}
		}
		select {
		case <-ctx.Done():
			s.logger.Warn("context cancelled, stopping block update listener")
			return
		case <-blockReadyChan:
			blockReadyCb()
			ticker.Reset(s.blockWaitTime)
		case <-ticker.C: // timeout, manually check for new blocks
			blockReadyCb()
		}
	}
}

func (ks *KarlsenApi) GetBlockTemplate(
	client *gostratum.StratumContext) (*appmessage.GetBlockTemplateResponseMessage, error) {
	template, err := ks.karlsend.GetBlockTemplate(client.WalletAddr,
		fmt.Sprintf(`'%s' via karlsen-network/karlsen-stratum-bridge_%s`, client.RemoteApp, version))
	if err != nil {
		return nil, errors.Wrap(err, "failed fetching new block template from karlsen")
	}
	return template, nil
}
