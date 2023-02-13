package helper

import (
	"fmt"
	"math"
)

type Matrix struct {
	//矩阵结构
	N, M int //m是列数,n是⾏数
	Data [][]float64
}

func (mat *Matrix) Set(m int, n int, data [][]float64) {
	//m是列数,n是⾏数,data是矩阵数据（从左到右由上到下填充）
	mat.M = m
	mat.N = n
	if len(data) == 0 || (len(data) != m || len(data[0]) != n) {
		err := fmt.Errorf("matrix size m, n does not match")
		panic(err)
	}
	mat.Data = data
}

func Mul(a Matrix, b Matrix) [][]float64 {
	if a.M == b.N {
		res := [][]float64{}
		// 以A行乘以B列
		for i := 0; i < a.N; i++ {
			t := []float64{}
			// a.data[i] 是第 i+1 行
			for j := 0; j < b.M; j++ {
				r := float64(0)
				for k := 0; k < a.M; k++ {
					r += a.Data[i][k] * b.Data[k][j]
				}
				t = append(t, r)
			}
			res = append(res, t)
		}
		return res
	} else {
		// "两矩阵⽆法进⾏相乘运算")
		return [][]float64{}
	}
}

func Transpose(a Matrix) Matrix {
	var b Matrix
	b.M = a.N
	b.N = a.M
	var res = [][]float64{}

	for i := 0; i < a.M; i++ {
		var t = []float64{}
		for j := 0; j < a.N; j++ {
			t = append(t, a.Data[j][i])
		}
		res = append(res, t)
	}
	b.Data = res
	return b
}

func Det(Matrix [][]float64, N int) float64 {
	var T0, T1, T2, Cha int
	var Num float64
	var B [][]float64

	if N > 0 {
		Cha = 0
		for i := 0; i < N; i++ {
			var tmpArr []float64
			for j := 0; j < N; j++ {
				tmpArr = append(tmpArr, 0)
			}
			B = append(B, tmpArr)
		}
		Num = 0
		for T0 = 0; T0 <= N; T0++ { //T0循环
			for T1 = 1; T1 <= N; T1++ { //T1循环
				for T2 = 0; T2 <= N-1; T2++ { //T2循环
					if T2 == T0 {
						Cha = 1
					}
					B[T1-1][T2] = Matrix[T1][T2+Cha]
				} //T2循环
				Cha = 0
			} //T1循环
			Num = Num + Matrix[0][T0]*Det(B, N-1)*math.Pow(-1, float64(T0))
		} //T0循环
		return Num
	} else if N == 0 {
		return Matrix[0][0]
	}
	return 0
}

func Inverse(S1 Matrix) (MatrixC [][]float64) {
	N := S1.N - 1
	Matrix := S1.Data
	var T0, T1, T2, T3 int
	var B [][]float64
	for i := 0; i < N; i++ {
		var tmpArr []float64
		for j := 0; j < N; j++ {
			tmpArr = append(tmpArr, 0)
		}
		B = append(B, tmpArr)
	}

	for i := 0; i < N+1; i++ {
		var tmpArr []float64
		for j := 0; j < N+1; j++ {
			tmpArr = append(tmpArr, 0)
		}
		MatrixC = append(MatrixC, tmpArr)
	}

	Chay := 0
	Chax := 0
	add := 1 / Det(Matrix, N)
	for T0 = 0; T0 <= N; T0++ {
		for T3 = 0; T3 <= N; T3++ {
			for T1 = 0; T1 <= N-1; T1++ {
				if T1 < T0 {
					Chax = 0
				} else {
					Chax = 1
				}
				for T2 = 0; T2 <= N-1; T2++ {
					if T2 < T3 {
						Chay = 0
					} else {
						Chay = 1
					}
					B[T1][T2] = Matrix[T1+Chax][T2+Chay]
				} //T2循环
			} //T1循环
			Det(B, N-1)
			MatrixC[T3][T0] = Det(B, N-1) * add * (math.Pow(-1, float64(T0+T3)))
		}
	}
	return MatrixC
}

type DualQuat struct {
	q [4]float64
	p [4]float64
}

func (dq *DualQuat) DualQuat2Pose() Matrix {
	mat := make([][]float64, 4)
	for i := 0; i < 4; i++ {
		mat[i] = make([]float64, 4)
	}
	ret := Matrix{4, 4, mat}
	mat[0][0] = dq.q[0]*dq.q[0] + dq.q[1]*dq.q[1] - dq.q[2]*dq.q[2] - dq.q[3]*dq.q[3]
	mat[0][1] = 2.0 * (dq.q[1]*dq.q[2] - dq.q[0]*dq.q[3])
	mat[0][2] = 2.0 * (dq.q[1]*dq.q[3] + dq.q[0]*dq.q[2])
	mat[0][3] = 2.0 * (dq.q[0]*dq.p[1] - dq.q[1]*dq.p[0] + dq.q[2]*dq.p[3] - dq.q[3]*dq.p[2])

	mat[1][0] = 2.0 * (dq.q[1]*dq.q[2] + dq.q[0]*dq.q[3])
	mat[1][1] = dq.q[0]*dq.q[0] - dq.q[1]*dq.q[1] + dq.q[2]*dq.q[2] - dq.q[3]*dq.q[3]
	mat[1][2] = 2.0 * (dq.q[2]*dq.q[3] - dq.q[0]*dq.q[1])
	mat[1][3] = 2.0 * (dq.q[0]*dq.p[2] - dq.q[1]*dq.p[3] - dq.q[2]*dq.p[0] + dq.q[3]*dq.p[1])

	mat[2][0] = 2.0 * (dq.q[1]*dq.q[3] - dq.q[0]*dq.q[2])
	mat[2][1] = 2.0 * (dq.q[2]*dq.q[3] + dq.q[0]*dq.q[1])
	mat[2][2] = dq.q[0]*dq.q[0] - dq.q[1]*dq.q[1] - dq.q[2]*dq.q[2] + dq.q[3]*dq.q[3]
	mat[2][3] = 2.0 * (dq.q[0]*dq.p[3] + dq.q[1]*dq.p[2] - dq.q[2]*dq.p[1] - dq.q[3]*dq.p[0])

	mat[3][0] = 0.0
	mat[3][1] = 0.0
	mat[3][2] = 0.0
	mat[3][3] = 1.0
	return ret
}
