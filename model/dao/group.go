package dao

import (
	"ichat-go/model/entity"
)

type GroupDao interface {
	CreateGroup(g *entity.Group)
	GetMemberUserIds(groupId uint64) []uint64
	CreateMember(g *entity.Group, userIds []uint64)
	FindGroupById(groupId uint64) *entity.Group
	FindGroups(groupIds []uint64) []*entity.Group
	GetMembers(groupId uint64) []*entity.User
}

type groupDao struct {
	tx Tx
}

func (d groupDao) CreateGroup(g *entity.Group) {
	assertNoError(d.tx.Create(g))
}

func (d groupDao) GetMemberUserIds(groupId uint64) []uint64 {
	var ids []uint64
	d.tx.Model(&entity.GroupMember{}).Where("group_id = ?", groupId).Pluck("user_id", &ids)
	return ids
}

func (d groupDao) CreateMember(g *entity.Group, userIds []uint64) {
	for _, userId := range userIds {
		m := &entity.GroupMember{
			GroupId: g.GroupId,
			UserId:  userId,
		}
		assertNoError(d.tx.Create(m))
	}
}

func (d groupDao) FindGroupById(groupId uint64) *entity.Group {
	var group entity.Group
	tx := d.tx.First(&group, groupId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &group
}

func (d groupDao) FindGroups(groupIds []uint64) []*entity.Group {
	var groups []*entity.Group
	d.tx.Find(&groups, groupIds)
	m := make(map[uint64]*entity.Group)
	for _, group := range groups {
		m[group.GroupId] = group
	}
	results := make([]*entity.Group, 0, len(groupIds))
	for _, groupId := range groupIds {
		group, _ := m[groupId]
		results = append(results, group)
	}
	return results
}

func (d groupDao) GetMembers(groupId uint64) []*entity.User {
	var members []*entity.User
	tx := d.tx.Model(&entity.GroupMember{}).Select("users.*").
		Joins("LEFT JOIN users ON group_members.user_id = users.user_id").
		Where("group_id = ?", groupId).
		Find(&members)
	if checkIsEmpty(tx) {
		return nil
	}
	return members
}

func NewGroupDao(tx Tx) GroupDao {
	return groupDao{tx: tx}
}
