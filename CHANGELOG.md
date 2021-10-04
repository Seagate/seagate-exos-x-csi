## [1.0.6](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.5...v1.0.6) (2021-10-04)

### Chores

- multi step build driver image step ([c64e84c](https://github.com/Seagate/seagate-exos-x-csi/commit/c64e84c160a068291ece316e23c1b6a3e3fb7659))

## [1.0.5](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.4...v1.0.5) (2021-10-04)

### Chores

- retrieve latest release for publish ([c904ea7](https://github.com/Seagate/seagate-exos-x-csi/commit/c904ea70d7db878dbc837ef8f6b2b55c5912d93f))

## [1.0.4](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.3...v1.0.4) (2021-10-04)

### Chores

- corrections to allow CD publish ([19a6267](https://github.com/Seagate/seagate-exos-x-csi/commit/19a6267f8b49e37518e209b050b07caf9f6d44e8))

### Other

- Merge branch 'main' of github.com:Seagate/seagate-exos-x-csi into main ([433eef0](https://github.com/Seagate/seagate-exos-x-csi/commit/433eef0b9217d0dbb367c7f75ad322d8ded0e7aa))

## [1.0.3](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.2...v1.0.3) (2021-10-04)

### Chores

- add publish workflow ([5c45073](https://github.com/Seagate/seagate-exos-x-csi/commit/5c450738b75e2b953a9653f3e6cba3ee9a5caafb))

### Other

- Merge branch 'main' of github.com:Seagate/seagate-exos-x-csi into main ([7374583](https://github.com/Seagate/seagate-exos-x-csi/commit/73745837bb422bf9c210f90ebdb9815d5546dde3))

## [1.0.2](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.1...v1.0.2) (2021-10-03)

### Chores

- trim changelog ([8ed43d2](https://github.com/Seagate/seagate-exos-x-csi/commit/8ed43d2bd2e0c06f1cd65834ac4fca1741895c84))

## [1.0.1](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.0...v1.0.1) (2021-10-02)

### Chores

- address ansi-regex security alert to use v5.0.1 ([5a1120b](https://github.com/Seagate/seagate-exos-x-csi/commit/5a1120b72e91bb7211a5d859c9eba87bb41cdf76))

# 1.0.0 (2021-10-02)

### Features

- **redhat-ubi:** Switched csi driver container image to use redhat/ubi8 ([1fa59fc](https://github.com/Seagate/seagate-exos-x-csi/commit/1fa59fc938ffe7eb1e98c6ae539baee44234a62a))

### Other

- removed unused portals type ([36a92d0](https://github.com/Seagate/seagate-exos-x-csi/commit/36a92d0ae3cf1429608e0b97afc5b9adf7d17ae1))
- Discover iSCSI IQN and Portals from storage appliance, not required in StorageClass ([14951fd](https://github.com/Seagate/seagate-exos-x-csi/commit/14951fdfd6186b50fbf151657ba69796d9c2a463))
- Added STX screenshots to documentation ([4bb93da](https://github.com/Seagate/seagate-exos-x-csi/commit/4bb93dafa2e754bb6a8b617b7331dd99361fcd84))
- Added usage example document and sample files. Updated README to remove reference to an OEM-specific configuration ([94fef7e](https://github.com/Seagate/seagate-exos-x-csi/commit/94fef7e5256520e0212b7c35f7ec21e81df0d734))
- update to latest sidecar versions ([a61e603](https://github.com/Seagate/seagate-exos-x-csi/commit/a61e603da7687fd94c1360e24de47e248b2b276f))
- updates to use older blkid options for compatibility reasons ([d50e30a](https://github.com/Seagate/seagate-exos-x-csi/commit/d50e30a230b31529763812586a67ce03f86f317f))
- storing devicePath, checking dependencies, and setting Multipath to true ([22668e0](https://github.com/Seagate/seagate-exos-x-csi/commit/22668e0a8309cc22033d88bf469efc9a78fac3e0))
- show snapshots working with all storage api versions, plus snapshot prefix ([284872b](https://github.com/Seagate/seagate-exos-x-csi/commit/284872b0ced8e0a9bd63676e9d0a67ca93791a01))
- show snapshots working with all storage api versions ([e9de7ce](https://github.com/Seagate/seagate-exos-x-csi/commit/e9de7ce2137a7dc820d210ade33b257ef1dcdf49))
- Update issue templates ([30e7ffc](https://github.com/Seagate/seagate-exos-x-csi/commit/30e7ffc6e4946bcf7c09f0b97ef2fb14e627113f))
- README updates ([05e66b7](https://github.com/Seagate/seagate-exos-x-csi/commit/05e66b711166c5818746a84e4cd41c5bdfafe2b9))
- bug/nodepath - required binaries now checked during driver init, iscsi clean up also ([a293808](https://github.com/Seagate/seagate-exos-x-csi/commit/a29380832323f170c2702d9ac577c86a313ab27b))
- bug/nodepath - path corrections for running iscsid/multipathd on host node ([6ef7bf9](https://github.com/Seagate/seagate-exos-x-csi/commit/6ef7bf99c1117c383a8427a87fe52e442488849b))
- fix(volume naming) added volPrefix to allow user defined volume prefix naming ([6cd34fa](https://github.com/Seagate/seagate-exos-x-csi/commit/6cd34fa47248d7ee7aa9a2cbd74dbb3792aedd0f))
- FMW-48421 - Added SecurityContextConstraints for OpenShift deployments ([697855c](https://github.com/Seagate/seagate-exos-x-csi/commit/697855cd5076fa2e101ae6256146c3efc5954026))
- Switched to ghcr.io/seagate/seagate-exos-x-csi ([c9d27ff](https://github.com/Seagate/seagate-exos-x-csi/commit/c9d27ff040ef95c97864cdea77367c1d0228a939))
- Removed iscsi and multipath sidecars ([bd149e0](https://github.com/Seagate/seagate-exos-x-csi/commit/bd149e0c745a5220669e6bb1b925b2ea9015b6ba))
- feat/rebrand additional mane changes to reduce use of exosx and other review changes ([2a38039](https://github.com/Seagate/seagate-exos-x-csi/commit/2a3803947ba492a1a499bef47ff48c0841ab26e3))
- localize to use github.com/Seagate ([f90c975](https://github.com/Seagate/seagate-exos-x-csi/commit/f90c975af5b3414a36e3e0ea686817f1570f2fe7))
