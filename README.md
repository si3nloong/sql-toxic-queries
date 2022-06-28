# Toxic Queries

> This is an experiment to test how slow can a toxic query be.

## Setup

<p>Machine: Macbook Pro</p>
<p>CPU: M1 Max</p>
<p>Memory: 32GB</p>
<p>OS: macOS Monterey</p>
<p>Record set: 200000</p>

## Benchmarks

| Statement              | Operation       |
| ---------------------- | --------------- |
| COUNT with \*          | 0.009784 ns/op  |
| COUNT with Primary Key | 0.01063 ns/op   |
| COUNT with Explain     | 0.0001766 ns/op |

| Statement              | Operation     |
| ---------------------- | ------------- |
| LIKE with Leading %    | 0.1275 ns/op  |
| LIKE without Leading % | 0.09251 ns/op |

| Statement               | Operation          |
| ----------------------- | ------------------ |
| Offset Based Pagination | 204339916750 ns/op |
| Cursor Based Pagination | 1339252750 ns/op   |

| Statement                    | Operation       |
| ---------------------------- | --------------- |
| INSERT with Stored Procedure | 0.0002593 ns/op |
| INSERT                       | 0.0001896 ns/op |
