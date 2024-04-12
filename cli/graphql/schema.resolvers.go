package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"fmt"
	"strconv"

	"numerous/cli/graphql/model"

	"github.com/99designs/gqlgen/graphql"
)

// OrganizationCreate is the resolver for the organizationCreate field.
func (r *mutationResolver) OrganizationCreate(ctx context.Context, input model.NewOrganization) (*model.Organization, error) {
	panic(fmt.Errorf("not implemented: OrganizationCreate - organizationCreate"))
}

// OrganizationRename is the resolver for the organizationRename field.
func (r *mutationResolver) OrganizationRename(ctx context.Context, organizationID string, name string) (model.OrganizationRenameResult, error) {
	panic(fmt.Errorf("not implemented: OrganizationRename - organizationRename"))
}

// OrganizationInvitationCreate is the resolver for the organizationInvitationCreate field.
func (r *mutationResolver) OrganizationInvitationCreate(ctx context.Context, organizationID string, input *model.OrganizationInvitationInput) (model.OrganizationInvitationCreateResult, error) {
	panic(fmt.Errorf("not implemented: OrganizationInvitationCreate - organizationInvitationCreate"))
}

// OrganizationInvitationAccept is the resolver for the organizationInvitationAccept field.
func (r *mutationResolver) OrganizationInvitationAccept(ctx context.Context, invitationID string) (model.OrganizationInvitationAcceptResult, error) {
	panic(fmt.Errorf("not implemented: OrganizationInvitationAccept - organizationInvitationAccept"))
}

// ToolCreate is the resolver for the toolCreate field.
func (r *mutationResolver) ToolCreate(ctx context.Context, input model.NewTool) (*model.Tool, error) {
	panic(fmt.Errorf("not implemented: ToolCreate - toolCreate"))
}

// ToolPublish is the resolver for the toolPublish field.
func (r *mutationResolver) ToolPublish(ctx context.Context, id string) (*model.Tool, error) {
	panic(fmt.Errorf("not implemented: ToolPublish - toolPublish"))
}

// ToolUnpublish is the resolver for the toolUnpublish field.
func (r *mutationResolver) ToolUnpublish(ctx context.Context, id string) (*model.Tool, error) {
	panic(fmt.Errorf("not implemented: ToolUnpublish - toolUnpublish"))
}

// ToolDelete is the resolver for the toolDelete field.
func (r *mutationResolver) ToolDelete(ctx context.Context, id string) (model.ToolDeleteResult, error) {
	panic(fmt.Errorf("not implemented: ToolDelete - toolDelete"))
}

// OrganizationAppDeploy is the resolver for the organizationAppDeploy field.
func (r *mutationResolver) OrganizationAppDeploy(ctx context.Context, appID string, organizationSlug string, appArchive graphql.Upload) (*model.AppDeploy, error) {
	panic(fmt.Errorf("not implemented: OrganizationAppDeploy - organizationAppDeploy"))
}

// JobStart is the resolver for the jobStart field.
func (r *mutationResolver) JobStart(ctx context.Context, toolHash string, hashType model.ToolHashType) (*model.Job, error) {
	panic(fmt.Errorf("not implemented: JobStart - jobStart"))
}

// JobStop is the resolver for the jobStop field.
func (r *mutationResolver) JobStop(ctx context.Context, id string) (*model.StopJobPayload, error) {
	panic(fmt.Errorf("not implemented: JobStop - jobStop"))
}

// ToolSessionCreate is the resolver for the toolSessionCreate field.
func (r *mutationResolver) ToolSessionCreate(ctx context.Context) (*model.ToolSession, error) {
	panic(fmt.Errorf("not implemented: ToolSessionCreate - toolSessionCreate"))
}

// ElementUpdate is the resolver for the elementUpdate field.
func (r *mutationResolver) ElementUpdate(ctx context.Context, toolSessionID string, clientID string, element model.ElementInput) (model.Element, error) {
	if convertedToolSessionID, err := strconv.ParseUint(toolSessionID, 10, 64); err != nil {
		return nil, err
	} else if result, err := r.ToolSessionService.UpdateElement(uint(convertedToolSessionID), clientID, ElementInputToDomain(element)); err != nil {
		return nil, err
	} else {
		return AppSessionElementFromDomain(result.Session, *result.Element), nil
	}
}

// ElementTrigger is the resolver for the elementTrigger field.
func (r *mutationResolver) ElementTrigger(ctx context.Context, toolSessionID string, clientID string, actionElementID string) (model.Element, error) {
	if convertedToolSessionID, err := strconv.ParseUint(toolSessionID, 10, 64); err != nil {
		return nil, err
	} else if result, err := r.ToolSessionService.TriggerAction(uint(convertedToolSessionID), clientID, actionElementID); err != nil {
		return nil, err
	} else {
		return AppSessionElementFromDomain(result.Session, *result.Element), nil
	}
}

// ElementSelectionUpdate is the resolver for the elementSelectionUpdate field.
func (r *mutationResolver) ElementSelectionUpdate(ctx context.Context, clientID string, elementSelection model.ElementSelectInput) (model.Element, error) {
	panic(fmt.Errorf("not implemented: ElementSelectionUpdate - elementSelectionUpdate"))
}

// ListElementAdd is the resolver for the listElementAdd field.
func (r *mutationResolver) ListElementAdd(ctx context.Context, clientID string, listElement *model.ListElementInput) (model.Element, error) {
	panic(fmt.Errorf("not implemented: ListElementAdd - listElementAdd"))
}

// ListElementRemove is the resolver for the listElementRemove field.
func (r *mutationResolver) ListElementRemove(ctx context.Context, clientID string, listItemID string) (model.Element, error) {
	panic(fmt.Errorf("not implemented: ListElementRemove - listElementRemove"))
}

// BuildPush is the resolver for the buildPush field.
func (r *mutationResolver) BuildPush(ctx context.Context, file graphql.Upload, id string) (*model.BuildConfiguration, error) {
	panic(fmt.Errorf("not implemented: BuildPush - buildPush"))
}

// Me is the resolver for the me field.
func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	panic(fmt.Errorf("not implemented: Me - me"))
}

// Organization is the resolver for the organization field.
func (r *queryResolver) Organization(ctx context.Context, organizationSlug *string) (model.OrganizationQueryResult, error) {
	panic(fmt.Errorf("not implemented: Organization - organization"))
}

// OrganizationInvitation is the resolver for the organizationInvitation field.
func (r *queryResolver) OrganizationInvitation(ctx context.Context, invitationID string) (model.OrganizationInvitationQueryResult, error) {
	panic(fmt.Errorf("not implemented: OrganizationInvitation - organizationInvitation"))
}

// PublicTools is the resolver for the publicTools field.
func (r *queryResolver) PublicTools(ctx context.Context) ([]*model.PublicTool, error) {
	panic(fmt.Errorf("not implemented: PublicTools - publicTools"))
}

// Tool is the resolver for the tool field.
func (r *queryResolver) Tool(ctx context.Context, id string) (*model.Tool, error) {
	panic(fmt.Errorf("not implemented: Tool - tool"))
}

// Tools is the resolver for the tools field.
func (r *queryResolver) Tools(ctx context.Context) ([]*model.Tool, error) {
	panic(fmt.Errorf("not implemented: Tools - tools"))
}

// Job is the resolver for the job field.
func (r *queryResolver) Job(ctx context.Context, id string) (*model.Job, error) {
	panic(fmt.Errorf("not implemented: Job - job"))
}

// JobsByTool is the resolver for the jobsByTool field.
func (r *queryResolver) JobsByTool(ctx context.Context, id string) ([]*model.Job, error) {
	panic(fmt.Errorf("not implemented: JobsByTool - jobsByTool"))
}

// Jobs is the resolver for the jobs field.
func (r *queryResolver) Jobs(ctx context.Context) ([]*model.Job, error) {
	panic(fmt.Errorf("not implemented: Jobs - jobs"))
}

// ToolSession is the resolver for the toolSession field.
func (r *queryResolver) ToolSession(ctx context.Context, id string) (*model.ToolSession, error) {
	toolSessionId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	toolSession, err := r.AppSessionsRepo.Read(uint(toolSessionId))
	if err != nil {
		return nil, err
	}

	s := AppSessionFromDomain(*toolSession)
	return s, nil
}

// ToolSessionEvent is the resolver for the toolSessionEvent field.
func (r *subscriptionResolver) ToolSessionEvent(ctx context.Context, toolSessionID string, clientID string) (<-chan model.ToolSessionEvent, error) {
	convertedToolSessionId, err := strconv.ParseUint(toolSessionID, 10, 64)
	if err != nil {
		return nil, err
	}

	subscription := make(chan model.ToolSessionEvent)
	session, err := r.AppSessionsRepo.Read(uint(convertedToolSessionId))
	if err != nil {
		return nil, err
	}

	events, err := r.ToolSessionService.Subscribe(ctx, toolSessionID, clientID)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			subscription <- AppSessionEventFromDomain(session, <-events)
		}
	}()

	return subscription, nil
}

// BuildEvents is the resolver for the buildEvents field.
func (r *subscriptionResolver) BuildEvents(ctx context.Context, buildID string, appPath *string) (<-chan model.BuildEvent, error) {
	panic(fmt.Errorf("not implemented: BuildEvents - buildEvents"))
}

// DeployEvents is the resolver for the deployEvents field.
func (r *subscriptionResolver) DeployEvents(ctx context.Context, toolID string) (<-chan model.BuildEvent, error) {
	panic(fmt.Errorf("not implemented: DeployEvents - deployEvents"))
}

// Logs is the resolver for the logs field.
func (r *subscriptionResolver) Logs(ctx context.Context, appID string) (<-chan *model.LogMessage, error) {
	panic(fmt.Errorf("not implemented: Logs - logs"))
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type (
	mutationResolver     struct{ *Resolver }
	queryResolver        struct{ *Resolver }
	subscriptionResolver struct{ *Resolver }
)
