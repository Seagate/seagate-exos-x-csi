## [1.5.9](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.8...v1.5.9) (2023-03-01)

### Bug Fixes

- Fix passing of FC initiators from node to controller ([9f9b6b9](https://github.com/Seagate/seagate-exos-x-csi/commit/9f9b6b93f93ecf07b8564edc93fefe0d36b5e7b3))

### Chores

- **deps:** bump golang.org/x/net ([f9107db](https://github.com/Seagate/seagate-exos-x-csi/commit/f9107db9795b74bf6cf8613d3db2a248ec75d183))
- **deps:** bump golang.org/x/text from 0.3.7 to 0.3.8 ([b85429b](https://github.com/Seagate/seagate-exos-x-csi/commit/b85429b2f2a33db39b47e55202cf5453bd1b6918))

### Documentation

- Remove defunct topology specification example ([a6e710c](https://github.com/Seagate/seagate-exos-x-csi/commit/a6e710c52dc9c927f65be886fe947f3bd386f47c))
- update multipath option to greedy ([e36a932](https://github.com/Seagate/seagate-exos-x-csi/commit/e36a932fe5387d96e032d1540ff83b65e313bb35))

## [1.5.8](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.7...v1.5.8) (2023-02-13)

### Bug Fixes

- Do not run filesystem checks if fs is already mounted ([70a5b71](https://github.com/Seagate/seagate-exos-x-csi/commit/70a5b71c4ab62a73414ab8fe8dc5abb252622203))
- Reduce maximum volume length to avoid copy volume truncation issue ([d769c22](https://github.com/Seagate/seagate-exos-x-csi/commit/d769c22dafd768ad22b79e3d7033046fb8e332e1))

### Chores

- **deps:** bump http-cache-semantics from 4.1.0 to 4.1.1 ([57445a2](https://github.com/Seagate/seagate-exos-x-csi/commit/57445a275abd245d7ccd73aadee7250d086a87bc))

### Other

- Update values.yaml ([220362b](https://github.com/Seagate/seagate-exos-x-csi/commit/220362b725dadf38432c0e35b8b01e3ec2675428))
- Update README.md to include more supported devices, and fix typos in docs/iscsi/multipath.conf ([6c2caa6](https://github.com/Seagate/seagate-exos-x-csi/commit/6c2caa6d93d8916e33f68eaff68589c83c402b73))
- Add a warning when volPrefix StorageClass param is too long and will be truncated ([2750fde](https://github.com/Seagate/seagate-exos-x-csi/commit/2750fdec13eb2aae8098cd20fb03d83132389260))
- switch to manual trigger ([73dde0d](https://github.com/Seagate/seagate-exos-x-csi/commit/73dde0dd0a7b20828b13c39eb5f924784767fe50))

## [1.5.7](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.6...v1.5.7) (2023-01-31)

### Bug Fixes

- squash a warning ([9453abb](https://github.com/Seagate/seagate-exos-x-csi/commit/9453abb114dc3299aa8813581ad54e14d31cf823))
- wait for semaphore in NodePublishVolume instead of requiring the CO to retry until the semaphore is free ([334cd0a](https://github.com/Seagate/seagate-exos-x-csi/commit/334cd0a7e66f428f01a9050ba27341e2fea1b53b))

### Other

- Merge pull request #60 from Seagate/bug#56 ([204b559](https://github.com/Seagate/seagate-exos-x-csi/commit/204b559bdaf864b741a3f7e7c71da4412881cace)), closes [#60](https://github.com/Seagate/seagate-exos-x-csi/issues/60) [Seagate/bug#56](https://github.com/Seagate/bug/issues/56) [Bug#56](https://github.com/Bug/issues/56)
- minor changes to improve the preflight and helm-package targets ([ed52e2a](https://github.com/Seagate/seagate-exos-x-csi/commit/ed52e2af848e790f478d7fd0a69c7c0bcdb45cc8))
- fix "could not parse topology requirements" error for iSCSI targets (#56, HS-332) ([13c749f](https://github.com/Seagate/seagate-exos-x-csi/commit/13c749feae805d8881f97ce088b229944b5ff00b)), closes [#56](https://github.com/Seagate/seagate-exos-x-csi/issues/56)

## [1.5.6](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.5...v1.5.6) (2023-01-20)

### Bug Fixes

- Fix XFS compatibility and return accessible topology ([3a43444](https://github.com/Seagate/seagate-exos-x-csi/commit/3a434441b007c0d5350affcb70d06968dc8df534))

### Other

- Merge pull request #59 from David-T-White/xfs-fscheck-fix ([1a0a321](https://github.com/Seagate/seagate-exos-x-csi/commit/1a0a321ca761b045e97d50b97c97acf5a1d7f1ff)), closes [#59](https://github.com/Seagate/seagate-exos-x-csi/issues/59)

## [1.5.5](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.4...v1.5.5) (2023-01-20)

### Chores

- Build OpenShift-compliant image by default. ([44180fe](https://github.com/Seagate/seagate-exos-x-csi/commit/44180fe13224d5567226d83fca2a5f0324912dfc))

## [1.5.4](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.3...v1.5.4) (2023-01-06)

### Bug Fixes

- Pass node ID with the topology map for initiator selection ([2d74b5e](https://github.com/Seagate/seagate-exos-x-csi/commit/2d74b5ef328dff245c0caec764650b9247aef79b))

## [1.5.3](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.2...v1.5.3) (2022-12-13)

### Chores

- update yaml values to correct version ([2041f81](https://github.com/Seagate/seagate-exos-x-csi/commit/2041f81f4a0fb7a3647e3d767cc56e6c0fa4205c))

### Other

- Hs 312/sas device not found (#53) ([b0dc972](https://github.com/Seagate/seagate-exos-x-csi/commit/b0dc972e54939077d724d40ee0c31d7965528b94)), closes [#53](https://github.com/Seagate/seagate-exos-x-csi/issues/53)

## [1.5.2](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.1...v1.5.2) (2022-10-21)

### Documentation

- SAS support documentation updates (#52) ([43ca0ff](https://github.com/Seagate/seagate-exos-x-csi/commit/43ca0fff64acd662776a970d8176994beffb990a)), closes [#52](https://github.com/Seagate/seagate-exos-x-csi/issues/52)

## [1.5.1](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.5.0...v1.5.1) (2022-10-18)

### Bug Fixes

- Add volume mount for controller run dir ([fae23f8](https://github.com/Seagate/seagate-exos-x-csi/commit/fae23f844b4383a30ca4a34180c2d68563913fac))

### Other

- Fix/sas initiator (#50) ([edcbd81](https://github.com/Seagate/seagate-exos-x-csi/commit/edcbd81f5d6276bcfcef2392849b0c9ea3b03852)), closes [#50](https://github.com/Seagate/seagate-exos-x-csi/issues/50)

# [1.5.0](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.4.0...v1.5.0) (2022-08-18)

### Chores

- Updated the version based on the new FC feature ([481a21b](https://github.com/Seagate/seagate-exos-x-csi/commit/481a21b80f04dc0aef02cae70b8b8ff35409383c))

### Features

- Fibre Channel interface layer support, leveraging SAS library ([92f00c6](https://github.com/Seagate/seagate-exos-x-csi/commit/92f00c680bdb5b819593c9487fcc6bf475929c65))

### Other

- Merge pull request #49 from Seagate/feat/fc ([12a5922](https://github.com/Seagate/seagate-exos-x-csi/commit/12a592255ef265e1ab56317e2564e48a8de8eab9)), closes [#49](https://github.com/Seagate/seagate-exos-x-csi/issues/49)

# [1.4.0](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.3.3...v1.4.0) (2022-08-05)

### Bug Fixes

- update version in helm and Makefile ([60a7542](https://github.com/Seagate/seagate-exos-x-csi/commit/60a75428df3ca9ca688a6a768e1b69ef25f29267))

### Chores

- upgrade Seagate/csi-lib-sas => v1.0.0 ([db1dca9](https://github.com/Seagate/seagate-exos-x-csi/commit/db1dca9845d91cb1d3d6bb1b85ab4cc850bc163c))

### Features

- adjust csi driver to match sas lin changes ([22c8a5f](https://github.com/Seagate/seagate-exos-x-csi/commit/22c8a5f8854dfa556a70db9ceeeea09f4e8c5759))
- SAS Connector name updates and comments ([1b7553c](https://github.com/Seagate/seagate-exos-x-csi/commit/1b7553cf27c3f444bfd33534775918293db58e64))

### Other

- Merge pull request #48 from Seagate/feat/saswwn ([335bd68](https://github.com/Seagate/seagate-exos-x-csi/commit/335bd68e2e0c98abdcd11dd7ab6c4734cba08861)), closes [#48](https://github.com/Seagate/seagate-exos-x-csi/issues/48)

## [1.3.3](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.3.2...v1.3.3) (2022-08-03)

### Chores

- update Makefile to support creating a signed helm package ([24d081b](https://github.com/Seagate/seagate-exos-x-csi/commit/24d081b51f3b965833d03500ef009e518e25b08f))

### Tests

- add option to specify imagePullSecrets (#47) ([b52c927](https://github.com/Seagate/seagate-exos-x-csi/commit/b52c927b6e8b69441ba1da0eaf6eed9a9f5a16a7)), closes [#47](https://github.com/Seagate/seagate-exos-x-csi/issues/47)

## [1.3.2](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.3.1...v1.3.2) (2022-08-03)

### Chores

- **deps-dev:** bump semantic-release from 18.0.1 to 19.0.3 ([864e291](https://github.com/Seagate/seagate-exos-x-csi/commit/864e29191daae0f92c2cd3a38c67a3c83e41f6d2))

### Other

- Merge pull request #44 from Seagate/dependabot/npm_and_yarn/semantic-release-19.0.3 ([d2b9847](https://github.com/Seagate/seagate-exos-x-csi/commit/d2b9847d033bf3e71625affec991ccffe1bb3068)), closes [#44](https://github.com/Seagate/seagate-exos-x-csi/issues/44)

## [1.3.1](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.3.0...v1.3.1) (2022-08-03)

### Chores

- **deps:** bump semver-regex from 3.1.3 to 3.1.4 ([f3686f3](https://github.com/Seagate/seagate-exos-x-csi/commit/f3686f308e882730a811dde824ae8668dd70759d))

### Other

- Merge pull request #43 from Seagate/dependabot/npm_and_yarn/semver-regex-3.1.4 ([e4d0b98](https://github.com/Seagate/seagate-exos-x-csi/commit/e4d0b981f4d41b5a93fbdd0450856038e1dd95f2)), closes [#43](https://github.com/Seagate/seagate-exos-x-csi/issues/43)
- Test/sanity crc (#35) ([3bd1fa8](https://github.com/Seagate/seagate-exos-x-csi/commit/3bd1fa8901a5720156caa5b0dc512699d56b94ac)), closes [#35](https://github.com/Seagate/seagate-exos-x-csi/issues/35)

# [1.3.0](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.2.3...v1.3.0) (2022-07-07)

### Features

- Add SAS support, NodePublishVolume added, using csi-lib-sas ([a542cdf](https://github.com/Seagate/seagate-exos-x-csi/commit/a542cdf01099b89bdcff33415c1c8958411d5146))
- Additional updates for SAS support ([38a2056](https://github.com/Seagate/seagate-exos-x-csi/commit/38a20561eff1176cc8680b444210d03b7db65646))
- Updated multipath.conf example in docs/sas ([d4ea31f](https://github.com/Seagate/seagate-exos-x-csi/commit/d4ea31fc5634cc5483edc7aec765533caf3dbb02))

### Other

- Merge pull request #46 from Seagate/feat/sas ([4cec69c](https://github.com/Seagate/seagate-exos-x-csi/commit/4cec69c9d5d123353c6ee90aa77a9b6f8a5e7f62)), closes [#46](https://github.com/Seagate/seagate-exos-x-csi/issues/46)

## [1.2.3](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.2.2...v1.2.3) (2022-06-01)

### Bug Fixes

- add thread safe mutex per volume publish/unpublish ([f1a2efd](https://github.com/Seagate/seagate-exos-x-csi/commit/f1a2efdd1255289e2a49b1aa1d76a87c96c48894))
- mutex protection, wwn for multipath confirmation ([c79a765](https://github.com/Seagate/seagate-exos-x-csi/commit/c79a76525ea6fc2a8244820e569982133aad6955))

### Other

- Merge pull request #42 from Seagate/fix/multipath ([6c99e54](https://github.com/Seagate/seagate-exos-x-csi/commit/6c99e5492d83b90ba6fd5c5ae133a47b45b8a6dd)), closes [#42](https://github.com/Seagate/seagate-exos-x-csi/issues/42)

## [1.2.2](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.2.1...v1.2.2) (2022-04-22)

### Bug Fixes

- **ListSnapshots:** Explicitly return empty snapshot response on -10058 rc ([bf37b55](https://github.com/Seagate/seagate-exos-x-csi/commit/bf37b551626f014e69c90789f7bdbd9033c6ff34))
- **ListSnapshots:** Handle -10058 return code when listing snapshots of non-existent volume ([04f14b5](https://github.com/Seagate/seagate-exos-x-csi/commit/04f14b59d962481922b5642ca174a6baa0179cf0))

### Other

- Merge pull request #41 from David-T-White/fix/listsnapshots ([a29e120](https://github.com/Seagate/seagate-exos-x-csi/commit/a29e120f1f5bb083f73f9a8d96bfdba4050a5ded)), closes [#41](https://github.com/Seagate/seagate-exos-x-csi/issues/41)
- Merge pull request #40 from David-T-White/bug/unmatched_quotes ([73b49f0](https://github.com/Seagate/seagate-exos-x-csi/commit/73b49f0c6b34ca7eb16e29e859dcc98d5f60ff3d)), closes [#40](https://github.com/Seagate/seagate-exos-x-csi/issues/40)
- Merge pull request #39 from Seagate/openshift-certification ([41a7b6f](https://github.com/Seagate/seagate-exos-x-csi/commit/41a7b6fd7466259cced34bb39a6923dc64632aa3)), closes [#39](https://github.com/Seagate/seagate-exos-x-csi/issues/39)
- add Dockerfile.redhat and license files for OpenShift certification ([75e86fc](https://github.com/Seagate/seagate-exos-x-csi/commit/75e86fcd107fbb1a1312c0791f57adc27dd96e94))
- fix mismatched quote errors from passing augmented volumeid ([e45263a](https://github.com/Seagate/seagate-exos-x-csi/commit/e45263afca9e6bfe0adbb046eac0976fa160424d))

## [1.2.1](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.2.0...v1.2.1) (2022-04-18)

### Bug Fixes

- Corrected Controller CreateVolume to use default iscsi when storage protocol is missing from StorageClass YAML ([8dd2d76](https://github.com/Seagate/seagate-exos-x-csi/commit/8dd2d765de4d36af46fa6c5032c7bcc74da218a9))
- Improved the storage protocol validation to exclude invalid strings ([67a9cde](https://github.com/Seagate/seagate-exos-x-csi/commit/67a9cdef117cc4c3bda72d506118213f14031cba))

### Other

- Merge pull request #37 from Seagate/fix/no-storage-protocol ([3162dd9](https://github.com/Seagate/seagate-exos-x-csi/commit/3162dd969a80640541ffb7b9305bac9f60567853)), closes [#37](https://github.com/Seagate/seagate-exos-x-csi/issues/37)

# [1.2.0](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.1.1...v1.2.0) (2022-04-07)

### Features

- **Storage:** Add use of augmented volume id to store volume name and storage protocol ([8355def](https://github.com/Seagate/seagate-exos-x-csi/commit/8355def4ecc13fef0209c59ca23c222f6fe1a8ce))
- **Storage:** Adding new storage package to handle iscsi, fc, and sas ([839b683](https://github.com/Seagate/seagate-exos-x-csi/commit/839b68350cb0ad12d324fc24f84c47b766f36eaf))
- **Storage:** iscsiNode implemented ([6255cc2](https://github.com/Seagate/seagate-exos-x-csi/commit/6255cc257f9a42d50c45c840d71e7bf192698853))
- **Storage:** Rename to Node ([6d4edb4](https://github.com/Seagate/seagate-exos-x-csi/commit/6d4edb488d9de5e55b9c8bcb9a3a49a7b1556f62))
- **Storage:** Spelling and naming corrections ([39e1709](https://github.com/Seagate/seagate-exos-x-csi/commit/39e17098ccaef5acf86bdbcf243cbddf292d79c1))
- **Storage:** Switch the blkid command timeout to a constant, with value of 10s ([ff70aef](https://github.com/Seagate/seagate-exos-x-csi/commit/ff70aefe83a59449a74265806836d578759116b4))
- **Storage:** Use iscsi for default storage protocol if not specified in StorageClass YAML ([340a694](https://github.com/Seagate/seagate-exos-x-csi/commit/340a6947c480d993ea0f852fca53681320aa2be4))

### Other

- Merge pull request #33 from Seagate/feat/storagei ([e33d829](https://github.com/Seagate/seagate-exos-x-csi/commit/e33d8299ca4b9172e3f6a4f0223290a5897d190f)), closes [#33](https://github.com/Seagate/seagate-exos-x-csi/issues/33)

## [1.1.1](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.1.0...v1.1.1) (2022-04-07)

### Chores

- **deps:** bump minimist from 1.2.5 to 1.2.6 ([35e59b3](https://github.com/Seagate/seagate-exos-x-csi/commit/35e59b3525b62b5352b68380466be05346af1a52))

### Other

- Merge pull request #34 from Seagate/dependabot/npm_and_yarn/minimist-1.2.6 ([93a0e5c](https://github.com/Seagate/seagate-exos-x-csi/commit/93a0e5cb565831b0b05bb96f25ddebc73be93e8b)), closes [#34](https://github.com/Seagate/seagate-exos-x-csi/issues/34)

# [1.1.0](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.11...v1.1.0) (2021-12-01)

### Bug Fixes

- code adjustments to pass csi-sanity ListSnapshots tests ([e2d4142](https://github.com/Seagate/seagate-exos-x-csi/commit/e2d41428cfc49bdb478037d5a33d11078a3cd9d5))
- Final changes for passing csi-sanity test suites ([e43f096](https://github.com/Seagate/seagate-exos-x-csi/commit/e43f096653b82ec2dfb740d3146a9ba66ae98b03))
- Final changes for passing csi-sanity test suites, v1.0.13 ([f328117](https://github.com/Seagate/seagate-exos-x-csi/commit/f3281171f752c3051dccda62b4432f8c764fb501))
- **workflow:** Corrections to workflow ([c6f4706](https://github.com/Seagate/seagate-exos-x-csi/commit/c6f470692edf564cf754201a102fd7ddb132b111))

### Chores

- publish helm package for seagate-exos-x-csi ([c3b9150](https://github.com/Seagate/seagate-exos-x-csi/commit/c3b9150e3832659a2fb2510924fde273b9b7b345))
- **release:** v1.0.12 ([b8051c7](https://github.com/Seagate/seagate-exos-x-csi/commit/b8051c7d18f9e3cbebf2b1a0903fd2b314ecf04d))
- **release:** v1.0.12 ([dacecb1](https://github.com/Seagate/seagate-exos-x-csi/commit/dacecb17665064bae204cd2cbedae00cdef5697b))
- **release:** v1.0.12 ([5cae3a9](https://github.com/Seagate/seagate-exos-x-csi/commit/5cae3a9b9d8779fb8f79b747f63cae05e31cfbfd))
- **release:** v1.0.12 ([0e7040c](https://github.com/Seagate/seagate-exos-x-csi/commit/0e7040c4700a3caccd03b7fcd70130b2f31d910c))
- **release:** v1.0.12 ([9e23ca5](https://github.com/Seagate/seagate-exos-x-csi/commit/9e23ca550233125c4381ab29a2e5128236497e45))

### Features

- csi-sanity passes ([1359168](https://github.com/Seagate/seagate-exos-x-csi/commit/1359168e8f88cc4001bf86a0751fd4abe3960054))

### Other

- Merge pull request #32 from Seagate/feat/sanity-passes ([432bc0b](https://github.com/Seagate/seagate-exos-x-csi/commit/432bc0bdf6d9b4d2a2adbdb7313a478668e11941)), closes [#32](https://github.com/Seagate/seagate-exos-x-csi/issues/32)
- Merge pull request #31 from Seagate/fix/workflow ([6136ff0](https://github.com/Seagate/seagate-exos-x-csi/commit/6136ff02c368c95a1f6ff27d639d88b8b479d238)), closes [#31](https://github.com/Seagate/seagate-exos-x-csi/issues/31)
- Merge pull request #30 from Seagate/test/csi-sanity-other ([e4472fa](https://github.com/Seagate/seagate-exos-x-csi/commit/e4472fa8d6eca5e391034d38c1084ad54c342378)), closes [#30](https://github.com/Seagate/seagate-exos-x-csi/issues/30)
- Merge pull request #29 from Seagate/test/csi-sanity-other ([bc5c7b6](https://github.com/Seagate/seagate-exos-x-csi/commit/bc5c7b6508654a7163a0907fd11ff015fb89ec22)), closes [#29](https://github.com/Seagate/seagate-exos-x-csi/issues/29)
- Merge pull request #28 from Seagate/test/csi-sanity-volumes ([0c300a2](https://github.com/Seagate/seagate-exos-x-csi/commit/0c300a2f6df80578209ede2558706f097c712a3d)), closes [#28](https://github.com/Seagate/seagate-exos-x-csi/issues/28)
- correct release version in versions.yaml and makefile ([afbb348](https://github.com/Seagate/seagate-exos-x-csi/commit/afbb348e63604eaab889c6d8f891d09c31b8cac9))

### Tests

- correct .gitignore ([afaacd1](https://github.com/Seagate/seagate-exos-x-csi/commit/afaacd194f3e6925b06fe4f013d1a4df9bfda3e8))
- correct ControllerPublishVolume csi-sanity issues ([53fa232](https://github.com/Seagate/seagate-exos-x-csi/commit/53fa232e95a3e36e3cf5e9c4ff4524044ad0ed52))
- correct CreateSnapshot csi-sanity issues ([2b2c2e0](https://github.com/Seagate/seagate-exos-x-csi/commit/2b2c2e01e32d4fbea8fd2568c79834f5e9064e2e))
- correct CreateVolume csi-sanity issues ([8818657](https://github.com/Seagate/seagate-exos-x-csi/commit/881865734f495e2aaa94f26668df4dfeb5e38f19))
- correct DeleteSnapshot csi-sanity issues ([c799e4c](https://github.com/Seagate/seagate-exos-x-csi/commit/c799e4c41a78130c045431a488eedb959f070755))
- csi-sanity minor test corrections ([3e9a8c0](https://github.com/Seagate/seagate-exos-x-csi/commit/3e9a8c03011bf9ecd40a7b6704a2bd46828c8a4f))

## [1.0.12](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.11...v1.0.12) (2021-12-01)

### Bug Fixes

- code adjustments to pass csi-sanity ListSnapshots tests ([e2d4142](https://github.com/Seagate/seagate-exos-x-csi/commit/e2d41428cfc49bdb478037d5a33d11078a3cd9d5))
- Final changes for passing csi-sanity test suites ([e43f096](https://github.com/Seagate/seagate-exos-x-csi/commit/e43f096653b82ec2dfb740d3146a9ba66ae98b03))
- Final changes for passing csi-sanity test suites, v1.0.13 ([f328117](https://github.com/Seagate/seagate-exos-x-csi/commit/f3281171f752c3051dccda62b4432f8c764fb501))
- **workflow:** Corrections to workflow ([c6f4706](https://github.com/Seagate/seagate-exos-x-csi/commit/c6f470692edf564cf754201a102fd7ddb132b111))

### Chores

- publish helm package for seagate-exos-x-csi ([c3b9150](https://github.com/Seagate/seagate-exos-x-csi/commit/c3b9150e3832659a2fb2510924fde273b9b7b345))
- **release:** v1.0.12 ([dacecb1](https://github.com/Seagate/seagate-exos-x-csi/commit/dacecb17665064bae204cd2cbedae00cdef5697b))
- **release:** v1.0.12 ([5cae3a9](https://github.com/Seagate/seagate-exos-x-csi/commit/5cae3a9b9d8779fb8f79b747f63cae05e31cfbfd))
- **release:** v1.0.12 ([0e7040c](https://github.com/Seagate/seagate-exos-x-csi/commit/0e7040c4700a3caccd03b7fcd70130b2f31d910c))
- **release:** v1.0.12 ([9e23ca5](https://github.com/Seagate/seagate-exos-x-csi/commit/9e23ca550233125c4381ab29a2e5128236497e45))

### Other

- Merge pull request #31 from Seagate/fix/workflow ([6136ff0](https://github.com/Seagate/seagate-exos-x-csi/commit/6136ff02c368c95a1f6ff27d639d88b8b479d238)), closes [#31](https://github.com/Seagate/seagate-exos-x-csi/issues/31)
- Merge pull request #30 from Seagate/test/csi-sanity-other ([e4472fa](https://github.com/Seagate/seagate-exos-x-csi/commit/e4472fa8d6eca5e391034d38c1084ad54c342378)), closes [#30](https://github.com/Seagate/seagate-exos-x-csi/issues/30)
- Merge pull request #29 from Seagate/test/csi-sanity-other ([bc5c7b6](https://github.com/Seagate/seagate-exos-x-csi/commit/bc5c7b6508654a7163a0907fd11ff015fb89ec22)), closes [#29](https://github.com/Seagate/seagate-exos-x-csi/issues/29)
- Merge pull request #28 from Seagate/test/csi-sanity-volumes ([0c300a2](https://github.com/Seagate/seagate-exos-x-csi/commit/0c300a2f6df80578209ede2558706f097c712a3d)), closes [#28](https://github.com/Seagate/seagate-exos-x-csi/issues/28)
- correct release version in versions.yaml and makefile ([afbb348](https://github.com/Seagate/seagate-exos-x-csi/commit/afbb348e63604eaab889c6d8f891d09c31b8cac9))

### Tests

- correct .gitignore ([afaacd1](https://github.com/Seagate/seagate-exos-x-csi/commit/afaacd194f3e6925b06fe4f013d1a4df9bfda3e8))
- correct ControllerPublishVolume csi-sanity issues ([53fa232](https://github.com/Seagate/seagate-exos-x-csi/commit/53fa232e95a3e36e3cf5e9c4ff4524044ad0ed52))
- correct CreateSnapshot csi-sanity issues ([2b2c2e0](https://github.com/Seagate/seagate-exos-x-csi/commit/2b2c2e01e32d4fbea8fd2568c79834f5e9064e2e))
- correct CreateVolume csi-sanity issues ([8818657](https://github.com/Seagate/seagate-exos-x-csi/commit/881865734f495e2aaa94f26668df4dfeb5e38f19))
- correct DeleteSnapshot csi-sanity issues ([c799e4c](https://github.com/Seagate/seagate-exos-x-csi/commit/c799e4c41a78130c045431a488eedb959f070755))
- csi-sanity minor test corrections ([3e9a8c0](https://github.com/Seagate/seagate-exos-x-csi/commit/3e9a8c03011bf9ecd40a7b6704a2bd46828c8a4f))

## [1.0.12](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.11...v1.0.12) (2021-11-30)

### Bug Fixes

- code adjustments to pass csi-sanity ListSnapshots tests ([e2d4142](https://github.com/Seagate/seagate-exos-x-csi/commit/e2d41428cfc49bdb478037d5a33d11078a3cd9d5))
- Final changes for passing csi-sanity test suites ([e43f096](https://github.com/Seagate/seagate-exos-x-csi/commit/e43f096653b82ec2dfb740d3146a9ba66ae98b03))
- Final changes for passing csi-sanity test suites, v1.0.13 ([f328117](https://github.com/Seagate/seagate-exos-x-csi/commit/f3281171f752c3051dccda62b4432f8c764fb501))

### Chores

- publish helm package for seagate-exos-x-csi ([c3b9150](https://github.com/Seagate/seagate-exos-x-csi/commit/c3b9150e3832659a2fb2510924fde273b9b7b345))
- **release:** v1.0.12 ([5cae3a9](https://github.com/Seagate/seagate-exos-x-csi/commit/5cae3a9b9d8779fb8f79b747f63cae05e31cfbfd))
- **release:** v1.0.12 ([0e7040c](https://github.com/Seagate/seagate-exos-x-csi/commit/0e7040c4700a3caccd03b7fcd70130b2f31d910c))
- **release:** v1.0.12 ([9e23ca5](https://github.com/Seagate/seagate-exos-x-csi/commit/9e23ca550233125c4381ab29a2e5128236497e45))

### Other

- Merge pull request #30 from Seagate/test/csi-sanity-other ([e4472fa](https://github.com/Seagate/seagate-exos-x-csi/commit/e4472fa8d6eca5e391034d38c1084ad54c342378)), closes [#30](https://github.com/Seagate/seagate-exos-x-csi/issues/30)
- Merge pull request #29 from Seagate/test/csi-sanity-other ([bc5c7b6](https://github.com/Seagate/seagate-exos-x-csi/commit/bc5c7b6508654a7163a0907fd11ff015fb89ec22)), closes [#29](https://github.com/Seagate/seagate-exos-x-csi/issues/29)
- Merge pull request #28 from Seagate/test/csi-sanity-volumes ([0c300a2](https://github.com/Seagate/seagate-exos-x-csi/commit/0c300a2f6df80578209ede2558706f097c712a3d)), closes [#28](https://github.com/Seagate/seagate-exos-x-csi/issues/28)
- correct release version in versions.yaml and makefile ([afbb348](https://github.com/Seagate/seagate-exos-x-csi/commit/afbb348e63604eaab889c6d8f891d09c31b8cac9))

### Tests

- correct .gitignore ([afaacd1](https://github.com/Seagate/seagate-exos-x-csi/commit/afaacd194f3e6925b06fe4f013d1a4df9bfda3e8))
- correct ControllerPublishVolume csi-sanity issues ([53fa232](https://github.com/Seagate/seagate-exos-x-csi/commit/53fa232e95a3e36e3cf5e9c4ff4524044ad0ed52))
- correct CreateSnapshot csi-sanity issues ([2b2c2e0](https://github.com/Seagate/seagate-exos-x-csi/commit/2b2c2e01e32d4fbea8fd2568c79834f5e9064e2e))
- correct CreateVolume csi-sanity issues ([8818657](https://github.com/Seagate/seagate-exos-x-csi/commit/881865734f495e2aaa94f26668df4dfeb5e38f19))
- correct DeleteSnapshot csi-sanity issues ([c799e4c](https://github.com/Seagate/seagate-exos-x-csi/commit/c799e4c41a78130c045431a488eedb959f070755))
- csi-sanity minor test corrections ([3e9a8c0](https://github.com/Seagate/seagate-exos-x-csi/commit/3e9a8c03011bf9ecd40a7b6704a2bd46828c8a4f))

## [1.0.12](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.11...v1.0.12) (2021-11-22)

### Bug Fixes

- code adjustments to pass csi-sanity ListSnapshots tests ([e2d4142](https://github.com/Seagate/seagate-exos-x-csi/commit/e2d41428cfc49bdb478037d5a33d11078a3cd9d5))

### Chores

- publish helm package for seagate-exos-x-csi ([c3b9150](https://github.com/Seagate/seagate-exos-x-csi/commit/c3b9150e3832659a2fb2510924fde273b9b7b345))
- **release:** v1.0.12 ([0e7040c](https://github.com/Seagate/seagate-exos-x-csi/commit/0e7040c4700a3caccd03b7fcd70130b2f31d910c))
- **release:** v1.0.12 ([9e23ca5](https://github.com/Seagate/seagate-exos-x-csi/commit/9e23ca550233125c4381ab29a2e5128236497e45))

### Other

- Merge pull request #29 from Seagate/test/csi-sanity-other ([bc5c7b6](https://github.com/Seagate/seagate-exos-x-csi/commit/bc5c7b6508654a7163a0907fd11ff015fb89ec22)), closes [#29](https://github.com/Seagate/seagate-exos-x-csi/issues/29)
- Merge pull request #28 from Seagate/test/csi-sanity-volumes ([0c300a2](https://github.com/Seagate/seagate-exos-x-csi/commit/0c300a2f6df80578209ede2558706f097c712a3d)), closes [#28](https://github.com/Seagate/seagate-exos-x-csi/issues/28)
- correct release version in versions.yaml and makefile ([afbb348](https://github.com/Seagate/seagate-exos-x-csi/commit/afbb348e63604eaab889c6d8f891d09c31b8cac9))

### Tests

- correct .gitignore ([afaacd1](https://github.com/Seagate/seagate-exos-x-csi/commit/afaacd194f3e6925b06fe4f013d1a4df9bfda3e8))
- correct ControllerPublishVolume csi-sanity issues ([53fa232](https://github.com/Seagate/seagate-exos-x-csi/commit/53fa232e95a3e36e3cf5e9c4ff4524044ad0ed52))
- correct CreateSnapshot csi-sanity issues ([2b2c2e0](https://github.com/Seagate/seagate-exos-x-csi/commit/2b2c2e01e32d4fbea8fd2568c79834f5e9064e2e))
- correct CreateVolume csi-sanity issues ([8818657](https://github.com/Seagate/seagate-exos-x-csi/commit/881865734f495e2aaa94f26668df4dfeb5e38f19))
- correct DeleteSnapshot csi-sanity issues ([c799e4c](https://github.com/Seagate/seagate-exos-x-csi/commit/c799e4c41a78130c045431a488eedb959f070755))
- csi-sanity minor test corrections ([3e9a8c0](https://github.com/Seagate/seagate-exos-x-csi/commit/3e9a8c03011bf9ecd40a7b6704a2bd46828c8a4f))

## [1.0.12](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.11...v1.0.12) (2021-10-22)

### Chores

- publish helm package for seagate-exos-x-csi ([c3b9150](https://github.com/Seagate/seagate-exos-x-csi/commit/c3b9150e3832659a2fb2510924fde273b9b7b345))
- **release:** v1.0.12 ([9e23ca5](https://github.com/Seagate/seagate-exos-x-csi/commit/9e23ca550233125c4381ab29a2e5128236497e45))

### Other

- Merge pull request #28 from Seagate/test/csi-sanity-volumes ([0c300a2](https://github.com/Seagate/seagate-exos-x-csi/commit/0c300a2f6df80578209ede2558706f097c712a3d)), closes [#28](https://github.com/Seagate/seagate-exos-x-csi/issues/28)
- correct release version in versions.yaml and makefile ([afbb348](https://github.com/Seagate/seagate-exos-x-csi/commit/afbb348e63604eaab889c6d8f891d09c31b8cac9))

### Tests

- correct .gitignore ([afaacd1](https://github.com/Seagate/seagate-exos-x-csi/commit/afaacd194f3e6925b06fe4f013d1a4df9bfda3e8))
- correct ControllerPublishVolume csi-sanity issues ([53fa232](https://github.com/Seagate/seagate-exos-x-csi/commit/53fa232e95a3e36e3cf5e9c4ff4524044ad0ed52))
- correct CreateSnapshot csi-sanity issues ([2b2c2e0](https://github.com/Seagate/seagate-exos-x-csi/commit/2b2c2e01e32d4fbea8fd2568c79834f5e9064e2e))
- correct CreateVolume csi-sanity issues ([8818657](https://github.com/Seagate/seagate-exos-x-csi/commit/881865734f495e2aaa94f26668df4dfeb5e38f19))
- correct DeleteSnapshot csi-sanity issues ([c799e4c](https://github.com/Seagate/seagate-exos-x-csi/commit/c799e4c41a78130c045431a488eedb959f070755))
- csi-sanity minor test corrections ([3e9a8c0](https://github.com/Seagate/seagate-exos-x-csi/commit/3e9a8c03011bf9ecd40a7b6704a2bd46828c8a4f))

## [1.0.12](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.11...v1.0.12) (2021-10-21)

### Chores

- publish helm package for seagate-exos-x-csi ([c3b9150](https://github.com/Seagate/seagate-exos-x-csi/commit/c3b9150e3832659a2fb2510924fde273b9b7b345))

### Other

- correct release version in versions.yaml and makefile ([afbb348](https://github.com/Seagate/seagate-exos-x-csi/commit/afbb348e63604eaab889c6d8f891d09c31b8cac9))

## [1.0.11](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.10...v1.0.11) (2021-10-07)

### Chores

- correct docker inspect step to use all lowercase for image name ([ade61cb](https://github.com/Seagate/seagate-exos-x-csi/commit/ade61cbec9b5deb5a6ea04ecb695b0966d8f2f85))

## [1.0.10](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.9...v1.0.10) (2021-10-07)

### Chores

- use all lowercase for image name ([8cb5fbd](https://github.com/Seagate/seagate-exos-x-csi/commit/8cb5fbdf6ebc412cbdb0241e407e0d59fcf5a569))

## [1.0.9](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.8...v1.0.9) (2021-10-07)

### Chores

- correct workflow error using run and uses ([4ca7c52](https://github.com/Seagate/seagate-exos-x-csi/commit/4ca7c523753e4d879a1b182be40ba119f3a0943b))

### Other

- Merge branch 'main' of github.com:Seagate/seagate-exos-x-csi into main ([d025e00](https://github.com/Seagate/seagate-exos-x-csi/commit/d025e0017deb5bbe1ab33da576fc85f337b57479))

## [1.0.8](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.7...v1.0.8) (2021-10-07)

### Chores

- correct version used for workflow Build and Push Docker Image, add Inspect step ([143fece](https://github.com/Seagate/seagate-exos-x-csi/commit/143fece6a6254c3fcdae65e7b6c2f214682ff875))

### Other

- Merge pull request #26 from Seagate/test/csi-sanity ([c0cdaff](https://github.com/Seagate/seagate-exos-x-csi/commit/c0cdaff63730a5a049f8b4b1fef121a7192e6d50)), closes [#26](https://github.com/Seagate/seagate-exos-x-csi/issues/26)

### Tests

- moved simple test app to its own folder ([bb1bd2d](https://github.com/Seagate/seagate-exos-x-csi/commit/bb1bd2dabaf65d3e21ed2aac00e03b01b074566d))
- new command line script for running csi-sanity ([8b7d14b](https://github.com/Seagate/seagate-exos-x-csi/commit/8b7d14b54e45530f731c65629c6fc1d74c6ecb5f))

## [1.0.7](https://github.com/Seagate/seagate-exos-x-csi/compare/v1.0.6...v1.0.7) (2021-10-04)

### Chores

- add login action ([8129575](https://github.com/Seagate/seagate-exos-x-csi/commit/8129575c49a2be158c5f0695ba9aec43ad907250))

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
