# Changelog

## [v1.10.3](https://github.com/k1LoW/donegroup/compare/v1.10.2...v1.10.3) - 2026-02-12
### Other Changes
- chore(deps): bump actions/checkout from 4 to 5 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/55
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/57
- chore: setup tagpr labels by @k1LoW in https://github.com/k1LoW/donegroup/pull/59
- chore(deps): bump Songmu/tagpr from 1.8.0 to 1.9.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/58
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/60
- chore(deps): bump the dependencies group across 1 directory with 2 updates by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/62
- chore(deps): bump Songmu/tagpr from 1.10.0 to 1.11.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/63
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/64
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/65
- chore(deps): bump Songmu/tagpr from 1.12.1 to 1.15.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/donegroup/pull/66
- fix: add mutex protection for cleanupGroups and fix context leak in AwaiterWithKey by @k1LoW in https://github.com/k1LoW/donegroup/pull/67
- fix: protect cleanupGroups slice read with mutex in WaitWithContextAndKey by @k1LoW in https://github.com/k1LoW/donegroup/pull/68

## [v1.10.2](https://github.com/k1LoW/donegroup/compare/v1.10.1...v1.10.2) - 2024-06-08
### Other Changes
- Use context.AfterFunc instead of go func() by @k1LoW in https://github.com/k1LoW/donegroup/pull/52

## [v1.10.1](https://github.com/k1LoW/donegroup/compare/v1.10.0...v1.10.1) - 2024-06-08
### Other Changes
- Refactor withDoneGroup by @k1LoW in https://github.com/k1LoW/donegroup/pull/51

## [v1.10.0](https://github.com/k1LoW/donegroup/compare/v1.9.0...v1.10.0) - 2024-06-05
### New Features 🎉
- Add `WithoutCancel` by @k1LoW in https://github.com/k1LoW/donegroup/pull/49

## [v1.9.0](https://github.com/k1LoW/donegroup/compare/v1.8.1...v1.9.0) - 2024-06-04
### Breaking Changes 🛠
- Change the behaviour of `donegroup.Cancel` significantly. by @k1LoW in https://github.com/k1LoW/donegroup/pull/47

## [v1.8.1](https://github.com/k1LoW/donegroup/compare/v1.8.0...v1.8.1) - 2024-06-02
### Fix bug 🐛
- Cleanup functions should be executed immediately when the context is done. by @k1LoW in https://github.com/k1LoW/donegroup/pull/44
### Other Changes
- doneGroup.ctxw is not used anymore, so remove it. by @k1LoW in https://github.com/k1LoW/donegroup/pull/45

## [v1.8.0](https://github.com/k1LoW/donegroup/compare/v1.7.0...v1.8.0) - 2024-06-02
### Breaking Changes 🛠
- If timeout is reached, it should not be waited for. by @k1LoW in https://github.com/k1LoW/donegroup/pull/39
- Functions registered in Cleanup no longer need to do context handling. by @k1LoW in https://github.com/k1LoW/donegroup/pull/42
### Other Changes
- Use context.WithoutCancel instead of context.Background in the donegroup package. by @k1LoW in https://github.com/k1LoW/donegroup/pull/41

## [v1.7.0](https://github.com/k1LoW/donegroup/compare/v1.6.0...v1.7.0) - 2024-06-02
### New Features 🎉
- Add `CancelWith*Cause` by @k1LoW in https://github.com/k1LoW/donegroup/pull/37
### Fix bug 🐛
- Fix doneGroup._ctx tree by @k1LoW in https://github.com/k1LoW/donegroup/pull/35
### Other Changes
- Use sync.WaitGroup instead of errgroup.Group by @k1LoW in https://github.com/k1LoW/donegroup/pull/32
- Remove unnecessary for loop by @k1LoW in https://github.com/k1LoW/donegroup/pull/33
- Use context.CancelCauseFunc by @k1LoW in https://github.com/k1LoW/donegroup/pull/34
- Fix cancel timing by @k1LoW in https://github.com/k1LoW/donegroup/pull/36

## [v1.6.0](https://github.com/k1LoW/donegroup/compare/v1.5.1...v1.6.0) - 2024-05-21
### New Features 🎉
- Add WithCancelCause by @k1LoW in https://github.com/k1LoW/donegroup/pull/27
- Add `WithDeadline` and `WithTimeout` by @k1LoW in https://github.com/k1LoW/donegroup/pull/30
### Other Changes
- chore(deps): bump golang.org/x/sync from 0.6.0 to 0.7.0 in the dependencies group by @dependabot in https://github.com/k1LoW/donegroup/pull/29
- chore(deps): bump actions/setup-go from 4 to 5 in the dependencies group by @dependabot in https://github.com/k1LoW/donegroup/pull/28

## [v1.5.1](https://github.com/k1LoW/donegroup/compare/v1.5.0...v1.5.1) - 2024-04-04

## [v1.5.0](https://github.com/k1LoW/donegroup/compare/v1.4.0...v1.5.0) - 2024-04-04
### New Features 🎉
- Add `donegroup.Go` by @k1LoW in https://github.com/k1LoW/donegroup/pull/23

## [v1.4.0](https://github.com/k1LoW/donegroup/compare/v1.3.0...v1.4.0) - 2024-02-07
### Breaking Changes 🛠
- Always execute all cleanup functions. by @k1LoW in https://github.com/k1LoW/donegroup/pull/21

## [v1.3.0](https://github.com/k1LoW/donegroup/compare/v1.2.0...v1.3.0) - 2024-02-07
### Breaking Changes 🛠
- Add Awaitable by @k1LoW in https://github.com/k1LoW/donegroup/pull/19
### Other Changes
- Add ErrNotContainDoneGroup by @k1LoW in https://github.com/k1LoW/donegroup/pull/17

## [v1.2.0](https://github.com/k1LoW/donegroup/compare/v1.1.0...v1.2.0) - 2024-02-07
### New Features 🎉
- Add Cancel for canceling context and waiting for cleanup functions at once. by @k1LoW in https://github.com/k1LoW/donegroup/pull/16

## [v1.1.0](https://github.com/k1LoW/donegroup/compare/v1.0.0...v1.1.0) - 2024-02-07
### New Features 🎉
- Add donegroup.Awaiter by @k1LoW in https://github.com/k1LoW/donegroup/pull/14
### Other Changes
- Add test for no cleanup functions by @k1LoW in https://github.com/k1LoW/donegroup/pull/12

## [v1.0.0](https://github.com/k1LoW/donegroup/compare/v0.2.3...v1.0.0) - 2024-02-06
### Other Changes
- Add example by @k1LoW in https://github.com/k1LoW/donegroup/pull/11

## [v0.2.3](https://github.com/k1LoW/donegroup/compare/v0.2.2...v0.2.3) - 2024-02-05
### Fix bug 🐛
- Fix typo by @k1LoW in https://github.com/k1LoW/donegroup/pull/9

## [v0.2.2](https://github.com/k1LoW/donegroup/compare/v0.2.1...v0.2.2) - 2024-02-05
### New Features 🎉
- Add WaitWithTimeout* by @k1LoW in https://github.com/k1LoW/donegroup/pull/7

## [v0.2.1](https://github.com/k1LoW/donegroup/compare/v0.2.0...v0.2.1) - 2024-02-05

## [v0.2.0](https://github.com/k1LoW/donegroup/compare/v0.1.0...v0.2.0) - 2024-02-05
### Breaking Changes 🛠
- Add WaitWithTimeout by @k1LoW in https://github.com/k1LoW/donegroup/pull/2
- Add WaitWithContext instead of WaitWithTimeout by @k1LoW in https://github.com/k1LoW/donegroup/pull/4

## [v0.0.1](https://github.com/k1LoW/donegroup/commits/v0.0.1) - 2024-02-05
