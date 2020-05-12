### note

- `AutoExpire bool` 的设置只在初次使用时候有效。当你需要运行时变更时，因为索引需要删除重建，所以这个过程不是自动化的，需要手动设置 mongo 的 `ExpiredAtField` 索引。
