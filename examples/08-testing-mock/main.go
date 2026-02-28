package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// 测试和 Mock 示例：展示如何在测试中使用依赖注入

type Database interface {
	Query(sql string) ([]map[string]interface{}, error)
}

type RealDatabase struct {
	DSN string
}

func (d *RealDatabase) Query(sql string) ([]map[string]interface{}, error) {
	// 实际数据库查询
	return nil, nil
}

type MockDatabase struct {
	Data []map[string]interface{}
}

func (d *MockDatabase) Query(sql string) ([]map[string]interface{}, error) {
	return d.Data, nil
}

type UserService struct {
	DB Database
}

func (s *UserService) GetUser(id int) (map[string]interface{}, error) {
	results, err := s.DB.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %d", id))
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return results[0], nil
	}
	return nil, nil
}

func main() {
	// 生产环境配置
	prodContainer := &godi.Container{}
	prodContainer.Add(godi.Provide(&RealDatabase{DSN: "mysql://localhost/prod"}))
	prodContainer.Add(godi.Provide(&UserService{DB: &RealDatabase{DSN: "mysql://localhost/prod"}}))

	// 测试环境配置（使用 Mock）
	testContainer := &godi.Container{}
	mockDB := &MockDatabase{
		Data: []map[string]interface{}{
			{"id": 1, "name": "测试用户", "email": "test@example.com"},
		},
	}
	testContainer.Add(godi.Provide(mockDB))
	testContainer.Add(godi.Provide(&UserService{DB: mockDB}))

	// 使用测试容器
	testUserSvc, _ := godi.Inject[*UserService](testContainer)
	user, _ := testUserSvc.GetUser(1)
	fmt.Printf("测试用户：%v\n", user)

	// 使用生产容器
	prodUserSvc, _ := godi.Inject[*UserService](prodContainer)
	fmt.Printf("生产服务就绪：%v\n", prodUserSvc != nil)

	fmt.Println("\n演示完成！")
	fmt.Println("- 测试环境使用 MockDatabase")
	fmt.Println("- 生产环境使用 RealDatabase")
}
