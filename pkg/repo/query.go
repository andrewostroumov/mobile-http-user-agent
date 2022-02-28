package repo

import "github.com/andrewostroumov/mobile-http-user-agent/pkg/repo/types"

func (r *Repo) RandDevice() (*types.Device, error) {
	var device types.Device

	row := r.pool.QueryRow(r.ctx, `SELECT devices.id, devices.build, devices.base_android_version FROM devices ORDER BY random() LIMIT 1`)

	err := row.Scan(&device.ID, &device.Build, &device.AndroidVersion)

	if err != nil {
		return nil, err
	}

	return &device, nil
}
