# Changelog

## [4.0.0](https://github.com/ngergs/websrv/compare/v3.5.6...v4.0.0) (2025-12-07)


### ⚠ BREAKING CHANGES

* server.RunTillWaitGroupFinishes has been replaced by a combination of the ListenGoServe method of server.Server and server.ShutdownAfterWaitGroup
* configurations have been moved from flags to config file / env vars. See updated README.md.
* Removed the following middlewares/handlers as provided by chi or the go stdlib (see refactored cmd/websrv/main.go): fileserver, gzip, requestId, timer. Changed the config-file setting `angular.csp-replace.file-name-regex` to `angular-csp-replace.file-path-regex`, the new format included a leading `/`.

### Features

* add a simple conditional healthcheck handler ([957241e](https://github.com/ngergs/websrv/commit/957241ee67a0e9bb1bd91d4e0925f6498432a5d9))
* add graceful shutdown ([75a6e03](https://github.com/ngergs/websrv/commit/75a6e03025759210dc5c21b6f68dbd89eb449804))
* add landlock (linux userspace sandbox feature) ([40fa2d7](https://github.com/ngergs/websrv/commit/40fa2d7d2bbb4b7d5533eeb46224fc0242476fda))
* add optional delay to the graceful shutdown ([eeb5350](https://github.com/ngergs/websrv/commit/eeb5350da0bf563b3668a61dfd384a5f8962b5be))
* call landlock after net.Listener have been setup ([f54a7d2](https://github.com/ngergs/websrv/commit/f54a7d2adce6494d3662d273590145fcbc5f0875))
* config refactoring ([bbf92f6](https://github.com/ngergs/websrv/commit/bbf92f617ee4e996fcd7bed60bef00a92da6ccfc))
* go version 1.18 ([f599e67](https://github.com/ngergs/websrv/commit/f599e672db60257f643520937709903ff4ce5e8a))
* go version 1.19 ([b3df275](https://github.com/ngergs/websrv/commit/b3df275b67c888b1362e51735fe7ba3915aa302a))
* h2c (unsupported http2) support ([bec1640](https://github.com/ngergs/websrv/commit/bec1640863cda235686a9ef9c32a66e33755b507))
* implement io.ReaderAt interface to in-memory fs ([451f1a9](https://github.com/ngergs/websrv/commit/451f1a9e68c5b215f6a6628925266b51c19f1069))
* initial release ([266f138](https://github.com/ngergs/websrv/commit/266f138909cd420d02f00b76abd0b658dd30a58d))
* listenGoServe method for server, simplify landlock network rules ([7314b74](https://github.com/ngergs/websrv/commit/7314b741a17e9161809858be785dded75d8de02b))
* optimize caching header handling ([989fca8](https://github.com/ngergs/websrv/commit/989fca88d1bb6b02d13fb5088709eb6002c2f09a))
* prometheus metrics ([2eec0eb](https://github.com/ngergs/websrv/commit/2eec0ebb6f3e7e8252222c0d9cbdd7b6ebea6378))
* prometheus metrics middleware and endpoint ([434de1c](https://github.com/ngergs/websrv/commit/434de1c14ff1902b39e4672e144ec2954a222fbc))
* read, write and tcp idle timeout ([7340b19](https://github.com/ngergs/websrv/commit/7340b199bd5b7ab544d63847206766ad2f112648))
* request rate limiting ([211fc29](https://github.com/ngergs/websrv/commit/211fc2947051bedae2f7f2ab3b044e0add6779be))
* stop immediately on second sigterm for graceful shutdown ([a965040](https://github.com/ngergs/websrv/commit/a9650408be3064658ef0c553a1b6f63447ac392a))
* use mutex free implementation for request-id and session-id ([4f89893](https://github.com/ngergs/websrv/commit/4f8989381a218dc3962f02ebcbc8e96f94d35e8c))
* use xsync for efficient concurrent maps ([22d3a26](https://github.com/ngergs/websrv/commit/22d3a262c4e59c0742ed0a5fe8a17f3998b06501))


### Bug Fixes

* add automemlimit ([1e0b259](https://github.com/ngergs/websrv/commit/1e0b259b44966fa3025b8f9e9e88c52d32e2f4ee))
* add default media type for .txt file extension ([9216020](https://github.com/ngergs/websrv/commit/9216020cf5f0e0134847cb79528f0f25838377e8))
* adjust go.mod for reproducible builds ([2ba400f](https://github.com/ngergs/websrv/commit/2ba400fa5c597b00d2227220c423d0425bc80558))
* also delay shutdown for health check ([a123bf5](https://github.com/ngergs/websrv/commit/a123bf58eaaaa998db8526f567a64300cf8598cd))
* context handling, lints ([8e64ac1](https://github.com/ngergs/websrv/commit/8e64ac1cc6e73abadfef4ee450c2ec15f76f7773))
* csp replace handler setup ([4f38179](https://github.com/ngergs/websrv/commit/4f38179c15cd2cd1ee8f95143163f9f2e33d7e34))
* dependency update ([ada799e](https://github.com/ngergs/websrv/commit/ada799ee3d66bae1a030bca4bd24a9ebe4f89ac4))
* dependency update ([4038c1a](https://github.com/ngergs/websrv/commit/4038c1a9a2f8f1854c78ed0060bf09a456c62759))
* dependency update ([a3cc15a](https://github.com/ngergs/websrv/commit/a3cc15ae043509c7e3d6fe8bdafa6c60b7fe75a1))
* dependency update ([73a72c4](https://github.com/ngergs/websrv/commit/73a72c46ebb942ee0b9019e4a05d46445e14a9e1))
* dependency updates ([6710b44](https://github.com/ngergs/websrv/commit/6710b44d00ee440124b8f4593b955a93098dc550))
* dependency updates ([fbdc0f8](https://github.com/ngergs/websrv/commit/fbdc0f8bb103937029b7493c14d628494e2a094b))
* dependency updates ([7fa75cf](https://github.com/ngergs/websrv/commit/7fa75cf13da5a66f59e3994d6d38a58d8edb3fdb))
* dependency updates ([41e0b26](https://github.com/ngergs/websrv/commit/41e0b26af339bca0928e4f6d3a1672b03d1d9b5b))
* dependency updates ([08ba7f5](https://github.com/ngergs/websrv/commit/08ba7f5cae03ff7c0c1d7b042088021111c53543))
* dependency updates ([4c69c31](https://github.com/ngergs/websrv/commit/4c69c3181950f19187a8e1f75d4302c59c61ebaf))
* dependency updates ([c89770b](https://github.com/ngergs/websrv/commit/c89770bd0a6b80d8ca8aa091a6d0181b6f02f835))
* dependency updates ([4b90307](https://github.com/ngergs/websrv/commit/4b90307991281356add48da974dcdc8b7d0363e6))
* dependency updates ([465a231](https://github.com/ngergs/websrv/commit/465a231dbbae9cf9fbcdce0893269f91cfac8177))
* dependency updates ([d503497](https://github.com/ngergs/websrv/commit/d503497e42e53cc3c08c110f72dc90df600e4382))
* dependency updates ([5f0a5ab](https://github.com/ngergs/websrv/commit/5f0a5ab2fc80d3db8d7cb05350729b634f066671))
* dependency updates ([3bf69c0](https://github.com/ngergs/websrv/commit/3bf69c01585d185efd723bf9b8f5ad85abc9fef2))
* dependency updates ([18b81ae](https://github.com/ngergs/websrv/commit/18b81ae4fcad085fefbf1615bf4bc9eaf76a8183))
* dependency updates ([651f189](https://github.com/ngergs/websrv/commit/651f18916918a5d4e7aff4e0f0153b09c14ac562))
* dependency updates ([a60eb2c](https://github.com/ngergs/websrv/commit/a60eb2ccfbd8d994a6673ffa955f9bc7a07a3f82))
* dependency updates ([a83e06d](https://github.com/ngergs/websrv/commit/a83e06de8ca627ec19fa14b60ea0243526fe7488))
* dependency updates ([5bd8825](https://github.com/ngergs/websrv/commit/5bd8825951caa6218d61afb334b7a86ca4ffdfb6))
* dependency updates ([f172ace](https://github.com/ngergs/websrv/commit/f172ace83cbef949fc2c2d97eba0527d12f2ae3d))
* dependency updates ([ef739eb](https://github.com/ngergs/websrv/commit/ef739eb473b18e05d1483373ed63373ba4191bda))
* dependency updates ([74b3b7c](https://github.com/ngergs/websrv/commit/74b3b7ccc556304e676c7da33530203fcab624c9))
* dependency updates ([7cf7b08](https://github.com/ngergs/websrv/commit/7cf7b08f8a939d6547f99b4cc5a39091f30e92b1))
* dependency updates ([41f0684](https://github.com/ngergs/websrv/commit/41f0684d7d3da08dfe0ce371a936ba45165bdf52))
* dependency updates ([ab7a006](https://github.com/ngergs/websrv/commit/ab7a006dea73c3858baf1ed4e869dec6c3c6bdba))
* dependency updates ([e1b54cf](https://github.com/ngergs/websrv/commit/e1b54cfb9e1cabf0eda5ee996a187f4964b1205d))
* dependency updates ([9579d6a](https://github.com/ngergs/websrv/commit/9579d6a918d103679f6f09ef66b2cd911c9c632a))
* dependency updates ([216a9bf](https://github.com/ngergs/websrv/commit/216a9bf96cbc7e610569695fb0ee67de1322062f))
* dependency updates ([4041e17](https://github.com/ngergs/websrv/commit/4041e1798bce79b11c2297971dea6e95abce16b1))
* dependency updates ([8120663](https://github.com/ngergs/websrv/commit/8120663ebfeaac80c5dc7ab1dac0f1a88fc9b0ba))
* dependency updates ([108081e](https://github.com/ngergs/websrv/commit/108081e1bf45f70d7527ce5c017da2f45e19bf61))
* dependency updates ([1d4ce76](https://github.com/ngergs/websrv/commit/1d4ce76e14ea52c7088b340dedc1fa829194361a))
* dependency updates ([1850d24](https://github.com/ngergs/websrv/commit/1850d24cee08b615a7441310a269da6f7d4b6941))
* dependency updates ([f632854](https://github.com/ngergs/websrv/commit/f63285417e90df5afd161fc6517b691997761d7f))
* dependency updates ([6ed4512](https://github.com/ngergs/websrv/commit/6ed45126ce7116b44a67c343b9e0128c80a4bbb2))
* dependency updates ([65db5ca](https://github.com/ngergs/websrv/commit/65db5cac266e074caa4b5af0dc03fd58b67f95fd))
* dependency updates ([3c6b742](https://github.com/ngergs/websrv/commit/3c6b742c10b5b655e22ad7ca06b9e0c86606b4e0))
* dependency updates ([28cc9c6](https://github.com/ngergs/websrv/commit/28cc9c6ac827d74b84f7a88702025938c520aa43))
* dependency updates ([356173c](https://github.com/ngergs/websrv/commit/356173c16d8a5628c8bbf0fd3485a97a4b760f73))
* dependency updates ([2f34db0](https://github.com/ngergs/websrv/commit/2f34db0a399a3baa2c4941579ebed3c509ad3a24))
* dependency updates ([7ea04ea](https://github.com/ngergs/websrv/commit/7ea04ea5110a5e67c151974180f08a4b34443c88))
* dependency updates ([2d545ae](https://github.com/ngergs/websrv/commit/2d545ae8e257ee21571c1cec0083a752c590ebb9))
* dependency updates ([e8b67c2](https://github.com/ngergs/websrv/commit/e8b67c252cd4cd6d6223d981a910c59597c21495))
* dependency updates ([6250f77](https://github.com/ngergs/websrv/commit/6250f77d3505020be0e2f8e6a09da05c0553e59d))
* dependency updates ([a17c161](https://github.com/ngergs/websrv/commit/a17c161fbeda8bb1ad3ae45bb305987302093382))
* dependency updates ([4350b6b](https://github.com/ngergs/websrv/commit/4350b6b22eb5cd6cb4ce70979faca4f201aef804))
* dependency updates ([b5cb363](https://github.com/ngergs/websrv/commit/b5cb3637f7b74de614823f8352623e5f3ca0b4b1))
* dependency updates ([f50a1dc](https://github.com/ngergs/websrv/commit/f50a1dcd8dcbb4ce425825d74cab66a2a226fea5))
* dependency updates ([a3ca9f2](https://github.com/ngergs/websrv/commit/a3ca9f271c96fc849f5de384d3d1e0e65d7472c9))
* dependency updates ([d187195](https://github.com/ngergs/websrv/commit/d187195afb83836747dccf818249d9865b4b65d6))
* dependency updates ([cbdc7ed](https://github.com/ngergs/websrv/commit/cbdc7ed7b43619442f10a8316907809200864d20))
* dependency updates ([b86496d](https://github.com/ngergs/websrv/commit/b86496d18ae592eee1abdccc33869fe8068b86ef))
* dependency updates ([e07fb52](https://github.com/ngergs/websrv/commit/e07fb526c2fb36cfca52be13f309b5c90b1b14bc))
* dependency updates, add automaxprocs ([e728e05](https://github.com/ngergs/websrv/commit/e728e054b1fb84a4f9e9ba04400ed8bec7f2d5fa))
* dockerfile (correct filename for goreleaser artifact) ([0f723f5](https://github.com/ngergs/websrv/commit/0f723f5e514fd2d141f27da811c2aa2f47d2452d))
* documentation ([d4da3a9](https://github.com/ngergs/websrv/commit/d4da3a921e02e99777a6dc63855ab7d909c804d7))
* documentation linting ([694a87b](https://github.com/ngergs/websrv/commit/694a87b0dfe51840945adb578d839d0dc4167a8f))
* documentation, small refactoring ([0891ee5](https://github.com/ngergs/websrv/commit/0891ee5e1b282fee8f780ca09413094af326737f))
* enable go 1.22 in go.mod ([6ad4960](https://github.com/ngergs/websrv/commit/6ad49604c8c16b59fb619b35d82f9e57416aa94a))
* error on config keys being used that do not match any actual config setting ([0a774f1](https://github.com/ngergs/websrv/commit/0a774f12a99cf3ed39c683981b2bed1bb392719a))
* fallback tof requested plain directories ([ab9f353](https://github.com/ngergs/websrv/commit/ab9f3538d66d156ba3be763e24afc867861ed694))
* field alignment optimization ([3f46fda](https://github.com/ngergs/websrv/commit/3f46fdac15b50a7ae5d83493e21f9027a6a9c0bc))
* gcp log format ([7fc9345](https://github.com/ngergs/websrv/commit/7fc9345389ce7a104d83684242d658bbe6866328))
* gcp log format for latency ([948c522](https://github.com/ngergs/websrv/commit/948c5224a1c78d6ad2b3dd1a460ad058c0aa2b4d))
* go patch update 1.19.1 ([49489ee](https://github.com/ngergs/websrv/commit/49489ee16812cbd2047c6d44d8a16e7f8cb48436))
* go patch version update ([a70ceae](https://github.com/ngergs/websrv/commit/a70ceae0d353a257b79b567242bacb28cfcce1bd))
* go v1.20.0 ([54c4b6c](https://github.com/ngergs/websrv/commit/54c4b6c4761ccb1ab92cd61d3534fe689f23fe6b))
* go v1.20.1 ([4c337ac](https://github.com/ngergs/websrv/commit/4c337ace85cfa1a39278633715ff6983827feae5))
* go version 1.19.5 ([ba46c07](https://github.com/ngergs/websrv/commit/ba46c07c3668bee8e01410cbcc81a32e741e740c))
* go version 1.20.3 ([c005aa9](https://github.com/ngergs/websrv/commit/c005aa99eb631e849a0b327625035e5651eb056e))
* go version 1.22.2 ([648ee9f](https://github.com/ngergs/websrv/commit/648ee9f1e084bcef782f4700a7f0533d1b3d01fb))
* go version update ([f54aaba](https://github.com/ngergs/websrv/commit/f54aabab88f4b1948617e77fdf46868c09924140))
* go-releaser inputs ([1371cf1](https://github.com/ngergs/websrv/commit/1371cf1a217f774e07e9602b02a2d078b4ca9ab9))
* golangci-lint false positives ([bd1be2e](https://github.com/ngergs/websrv/commit/bd1be2ea7ba71f3288a7d4d3439cf65138a28e6e))
* golangci-lint fixes ([e3acb46](https://github.com/ngergs/websrv/commit/e3acb466e838e3d5c08b540f090a3ce36810e3c0))
* goreleaser docker tag typo ([55fd2db](https://github.com/ngergs/websrv/commit/55fd2db1f56f950d7cda985cfaf12067f342b299))
* goreleaser full version tag typo ([56eda6a](https://github.com/ngergs/websrv/commit/56eda6aeecdcfd4139bc94c1ea28784d4ba3ad3e))
* gzip default compression level ([e3f4897](https://github.com/ngergs/websrv/commit/e3f4897a80261e2a6b12da5003ab24a5ceb80f24))
* gzip encoding media type handling ([f6c01e5](https://github.com/ngergs/websrv/commit/f6c01e5b381fad9f0fffe609f7c19e9b2ad2f65c))
* gzip setup ([5d0ed5d](https://github.com/ngergs/websrv/commit/5d0ed5dc5dc931f8190ae30662fff9e9e167101d))
* h2c Alt-Svc HTTP header ([4563753](https://github.com/ngergs/websrv/commit/45637535910b4f5bc95acb2ba87e52852844d0e5))
* handle graceful shutdown error ([d85ea69](https://github.com/ngergs/websrv/commit/d85ea69bd4f5e21737177054d513561652eb4d3f))
* header handling ([7238c8a](https://github.com/ngergs/websrv/commit/7238c8a46f33c571d69717bac54d8e90a557e26d))
* healthcheck test, bugfix ([09906bf](https://github.com/ngergs/websrv/commit/09906bfd5b0963dc0f9ac69d149682d0ba43ca59))
* implement io.Seeker for in memory files ([aa4fef8](https://github.com/ngergs/websrv/commit/aa4fef83515b5da682500f85a697d4325b25c5a4))
* include scheme and host in access log ([154ceda](https://github.com/ngergs/websrv/commit/154ceda68677cf4c051549384ceaf20e1017d974))
* landlock os.args error (tries to lock on wrong path) ([432eda9](https://github.com/ngergs/websrv/commit/432eda995866b6db078fd31718cfb893720b1668))
* mapstructure breaking change in koanf ([5f2ebb8](https://github.com/ngergs/websrv/commit/5f2ebb86618ed70fb7bda8a2cf773b823870a0c3))
* metrics/metrics-access-log difference corrected ([425df8e](https://github.com/ngergs/websrv/commit/425df8ecd6890b8607af475a49d3c0ab4a9fbd1f))
* more graceful shutdown on second SIGTERM ([c95a7d3](https://github.com/ngergs/websrv/commit/c95a7d3c4dd65f18628092fcfce577f95a86a9be))
* optimizations ([3a1aca0](https://github.com/ngergs/websrv/commit/3a1aca01dd5d7505b957cb6232b023f14500c78d))
* print version information at startup ([a15c719](https://github.com/ngergs/websrv/commit/a15c7196e2bb2816068dfda29d36a9134a5465b5))
* refactor promtheus types registration ([d8765d2](https://github.com/ngergs/websrv/commit/d8765d2473e23a9caa0691edd0fbcb524a6b92eb))
* renamen env vars ([a2bed9e](https://github.com/ngergs/websrv/commit/a2bed9e91236cdd45ca40a964cfd651e4f2107c8))
* set content-type on first csp-replace call ([77e0541](https://github.com/ngergs/websrv/commit/77e05410c21723f97262e6c4edce1b1fa7834b0d))
* simplify caching handler ([6264456](https://github.com/ngergs/websrv/commit/626445692b43f9cd122e0cc99f94bbdb47794db4))
* simplify config handling ([b7c4165](https://github.com/ngergs/websrv/commit/b7c4165f79df5303f4b005ee3222546b82a04cd5))
* stop health server only after all other servers are finished with their graceful shutdown ([25ac48f](https://github.com/ngergs/websrv/commit/25ac48f65a82cdc6728918819e497756131dce08))
* support multiple instanciations of the access metrics middleware ([0038519](https://github.com/ngergs/websrv/commit/00385196221e59ab2e642cc77e3fc0e85f6328f3))
* update distroless image ([6fc2f07](https://github.com/ngergs/websrv/commit/6fc2f073a6bf5524c9aa1ec0feb43b40d5d59e5d))
* use counter gauge vor egress_bytes_send ([97ef14b](https://github.com/ngergs/websrv/commit/97ef14bd823afee5e43de0000ef91fea1573b459))
* use math/rand/v2 ([77823c3](https://github.com/ngergs/websrv/commit/77823c345d9e166262ec448174337d075ea42053))
* use metrics-access-log flag to determine whether to add the middleware ([1acdd3c](https://github.com/ngergs/websrv/commit/1acdd3c903c11864a06507dd3164727e5e08c5e3))
* use metricsAccessLog flag for metrics middleware setup ([b9d08fc](https://github.com/ngergs/websrv/commit/b9d08fc780fb4a57043ffed933143aa919533707))
* use strings.Builder for url building, format fix ([07d9390](https://github.com/ngergs/websrv/commit/07d9390463c4be3003a7961e5ef49792a5184137))
* use sync.Map instead of sync.RWMutex ([7b17d31](https://github.com/ngergs/websrv/commit/7b17d31df10cce7eeb8052c3f775632dff628aa1))
* use type-safe map for cspReplace ([2800bb9](https://github.com/ngergs/websrv/commit/2800bb9fca4809ed201320cfa0db5cf957a70c40))
* use uint16 for ports (instead of int) ([092aba5](https://github.com/ngergs/websrv/commit/092aba57fb706dd7adefd43260dc07cd4347fc6c))
* use zerolog also for stdlib logs ([e648104](https://github.com/ngergs/websrv/commit/e64810422bbe2621d79d762a29bda4bf62397eee))
* using chi to simplify setup ([48f353e](https://github.com/ngergs/websrv/commit/48f353e608e7884521b8a76f520a08ae91b03314))

## [3.5.6](https://github.com/ngergs/websrv/compare/v3.5.5...v3.5.6) (2025-12-07)


### Bug Fixes

* context handling, lints ([8e64ac1](https://github.com/ngergs/websrv/commit/8e64ac1cc6e73abadfef4ee450c2ec15f76f7773))
* dependency updates ([6710b44](https://github.com/ngergs/websrv/commit/6710b44d00ee440124b8f4593b955a93098dc550))
* dependency updates ([fbdc0f8](https://github.com/ngergs/websrv/commit/fbdc0f8bb103937029b7493c14d628494e2a094b))

## [3.5.5](https://github.com/ngergs/websrv/compare/v3.5.4...v3.5.5) (2025-08-16)


### Bug Fixes

* dependency updates ([7fa75cf](https://github.com/ngergs/websrv/commit/7fa75cf13da5a66f59e3994d6d38a58d8edb3fdb))

## [3.5.4](https://github.com/ngergs/websrv/compare/v3.5.3...v3.5.4) (2025-08-05)


### Bug Fixes

* dependency updates ([41e0b26](https://github.com/ngergs/websrv/commit/41e0b26af339bca0928e4f6d3a1672b03d1d9b5b))

## [3.5.3](https://github.com/ngergs/websrv/compare/v3.5.2...v3.5.3) (2025-07-17)


### Bug Fixes

* dependency updates ([08ba7f5](https://github.com/ngergs/websrv/commit/08ba7f5cae03ff7c0c1d7b042088021111c53543))

## [3.5.2](https://github.com/ngergs/websrv/compare/v3.5.1...v3.5.2) (2025-06-27)


### Bug Fixes

* dependency updates ([4c69c31](https://github.com/ngergs/websrv/commit/4c69c3181950f19187a8e1f75d4302c59c61ebaf))

## [3.5.1](https://github.com/ngergs/websrv/compare/v3.5.0...v3.5.1) (2025-05-02)


### Bug Fixes

* dependency updates ([c89770b](https://github.com/ngergs/websrv/commit/c89770bd0a6b80d8ca8aa091a6d0181b6f02f835))

## [3.5.0](https://github.com/ngergs/websrv/compare/v3.4.1...v3.5.0) (2025-04-15)


### Features

* request rate limiting ([211fc29](https://github.com/ngergs/websrv/commit/211fc2947051bedae2f7f2ab3b044e0add6779be))


### Bug Fixes

* dependency updates ([4b90307](https://github.com/ngergs/websrv/commit/4b90307991281356add48da974dcdc8b7d0363e6))
* dependency updates ([465a231](https://github.com/ngergs/websrv/commit/465a231dbbae9cf9fbcdce0893269f91cfac8177))
* dependency updates ([d503497](https://github.com/ngergs/websrv/commit/d503497e42e53cc3c08c110f72dc90df600e4382))

## [3.4.1](https://github.com/ngergs/websrv/compare/v3.4.0...v3.4.1) (2025-02-13)


### Bug Fixes

* dependency updates ([5f0a5ab](https://github.com/ngergs/websrv/commit/5f0a5ab2fc80d3db8d7cb05350729b634f066671))
* dependency updates ([3bf69c0](https://github.com/ngergs/websrv/commit/3bf69c01585d185efd723bf9b8f5ad85abc9fef2))
* dependency updates ([18b81ae](https://github.com/ngergs/websrv/commit/18b81ae4fcad085fefbf1615bf4bc9eaf76a8183))
* dependency updates ([651f189](https://github.com/ngergs/websrv/commit/651f18916918a5d4e7aff4e0f0153b09c14ac562))
* dependency updates ([a60eb2c](https://github.com/ngergs/websrv/commit/a60eb2ccfbd8d994a6673ffa955f9bc7a07a3f82))

## [3.4.0](https://github.com/ngergs/websrv/compare/v3.3.3...v3.4.0) (2024-10-14)


### Features

* listenGoServe method for server, simplify landlock network rules ([7314b74](https://github.com/ngergs/websrv/commit/7314b741a17e9161809858be785dded75d8de02b))


### Bug Fixes

* add automemlimit ([1e0b259](https://github.com/ngergs/websrv/commit/1e0b259b44966fa3025b8f9e9e88c52d32e2f4ee))
* dependency updates ([a83e06d](https://github.com/ngergs/websrv/commit/a83e06de8ca627ec19fa14b60ea0243526fe7488))
* dependency updates ([5bd8825](https://github.com/ngergs/websrv/commit/5bd8825951caa6218d61afb334b7a86ca4ffdfb6))
* dependency updates ([f172ace](https://github.com/ngergs/websrv/commit/f172ace83cbef949fc2c2d97eba0527d12f2ae3d))
* dependency updates ([ef739eb](https://github.com/ngergs/websrv/commit/ef739eb473b18e05d1483373ed63373ba4191bda))
* dependency updates ([74b3b7c](https://github.com/ngergs/websrv/commit/74b3b7ccc556304e676c7da33530203fcab624c9))
* update distroless image ([6fc2f07](https://github.com/ngergs/websrv/commit/6fc2f073a6bf5524c9aa1ec0feb43b40d5d59e5d))

## [3.3.3](https://github.com/ngergs/websrv/compare/v3.3.2...v3.3.3) (2024-06-09)


### Bug Fixes

* dependency updates ([7cf7b08](https://github.com/ngergs/websrv/commit/7cf7b08f8a939d6547f99b4cc5a39091f30e92b1))
* dependency updates ([41f0684](https://github.com/ngergs/websrv/commit/41f0684d7d3da08dfe0ce371a936ba45165bdf52))

## [3.3.2](https://github.com/ngergs/websrv/compare/v3.3.1...v3.3.2) (2024-04-30)


### Bug Fixes

* dependency updates ([ab7a006](https://github.com/ngergs/websrv/commit/ab7a006dea73c3858baf1ed4e869dec6c3c6bdba))

## [3.3.1](https://github.com/ngergs/websrv/compare/v3.3.0...v3.3.1) (2024-04-11)


### Bug Fixes

* landlock os.args error (tries to lock on wrong path) ([432eda9](https://github.com/ngergs/websrv/commit/432eda995866b6db078fd31718cfb893720b1668))

## [3.3.0](https://github.com/ngergs/websrv/compare/v3.2.0...v3.3.0) (2024-04-11)


### Features

* use xsync for efficient concurrent maps ([22d3a26](https://github.com/ngergs/websrv/commit/22d3a262c4e59c0742ed0a5fe8a17f3998b06501))


### Bug Fixes

* dependency updates ([e1b54cf](https://github.com/ngergs/websrv/commit/e1b54cfb9e1cabf0eda5ee996a187f4964b1205d))
* dependency updates ([9579d6a](https://github.com/ngergs/websrv/commit/9579d6a918d103679f6f09ef66b2cd911c9c632a))
* simplify config handling ([b7c4165](https://github.com/ngergs/websrv/commit/b7c4165f79df5303f4b005ee3222546b82a04cd5))
* use math/rand/v2 ([77823c3](https://github.com/ngergs/websrv/commit/77823c345d9e166262ec448174337d075ea42053))

## [3.2.0](https://github.com/ngergs/websrv/compare/v3.1.7...v3.2.0) (2024-04-04)


### Features

* add landlock (linux userspace sandbox feature) ([40fa2d7](https://github.com/ngergs/websrv/commit/40fa2d7d2bbb4b7d5533eeb46224fc0242476fda))


### Bug Fixes

* dependency updates ([216a9bf](https://github.com/ngergs/websrv/commit/216a9bf96cbc7e610569695fb0ee67de1322062f))
* use uint16 for ports (instead of int) ([092aba5](https://github.com/ngergs/websrv/commit/092aba57fb706dd7adefd43260dc07cd4347fc6c))

## [3.1.7](https://github.com/ngergs/websrv/compare/v3.1.6...v3.1.7) (2024-04-03)


### Bug Fixes

* goreleaser full version tag typo ([56eda6a](https://github.com/ngergs/websrv/commit/56eda6aeecdcfd4139bc94c1ea28784d4ba3ad3e))

## [3.1.6](https://github.com/ngergs/websrv/compare/v3.1.5...v3.1.6) (2024-04-03)


### Bug Fixes

* goreleaser docker tag typo ([55fd2db](https://github.com/ngergs/websrv/commit/55fd2db1f56f950d7cda985cfaf12067f342b299))

## [3.1.5](https://github.com/ngergs/websrv/compare/v3.1.4...v3.1.5) (2024-04-03)


### Bug Fixes

* dependency updates ([4041e17](https://github.com/ngergs/websrv/commit/4041e1798bce79b11c2297971dea6e95abce16b1))
* go version 1.22.2 ([648ee9f](https://github.com/ngergs/websrv/commit/648ee9f1e084bcef782f4700a7f0533d1b3d01fb))

## [3.1.4](https://github.com/ngergs/websrv/compare/v3.1.3...v3.1.4) (2024-04-03)


### Bug Fixes

* dockerfile (correct filename for goreleaser artifact) ([0f723f5](https://github.com/ngergs/websrv/commit/0f723f5e514fd2d141f27da811c2aa2f47d2452d))

## [3.1.3](https://github.com/ngergs/websrv/compare/v3.1.2...v3.1.3) (2024-04-03)


### Bug Fixes

* adjust go.mod for reproducible builds ([2ba400f](https://github.com/ngergs/websrv/commit/2ba400fa5c597b00d2227220c423d0425bc80558))

## [3.1.2](https://github.com/ngergs/websrv/compare/v3.1.1...v3.1.2) (2024-04-03)


### Bug Fixes

* go-releaser inputs ([1371cf1](https://github.com/ngergs/websrv/commit/1371cf1a217f774e07e9602b02a2d078b4ca9ab9))

## [3.1.1](https://github.com/ngergs/websrv/compare/v3.1.0...v3.1.1) (2024-04-03)


### Bug Fixes

* golangci-lint false positives ([bd1be2e](https://github.com/ngergs/websrv/commit/bd1be2ea7ba71f3288a7d4d3439cf65138a28e6e))
